/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/steveredden/KindredCard/docs"
	"github.com/steveredden/KindredCard/internal/carddav"
	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/handlers"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/scheduler"
)

var ReleaseVersion = "v0.0.0-dev"

//	@title			KindredCard API
//	@version		1.0
//	@description	Personal CRM for managing contacts, relationships, and important dates

//	@contact.name	API Support
//	@contact.url	https://github.com/steveredden/KindredCard/issues

//	@license.name	AGPLv3
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.txt

//	@securityDefinitions.apikey	ApiTokenAuth
//	@in							header
//	@name						session
//	@description				API token for authentication

func main() {
	// Configuration from environment variables
	// logging level defined by OS ENV LOG_LEVEL
	logger.Init()

	appKey := getEnv("APP_KEY", "missing")
	if appKey == "missing" {
		logger.Fatal("[APP] An APP_KEY is required")
	}

	enableTwoWayCardDAV := (strings.ToUpper(getEnv("ENABLE_TWO_WAY_CARDDAV", "FALSE")) == "TRUE")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "kindredcard")
	dbPassword := getEnv("DB_PASSWORD", "kindredcardsecretpassword")
	dbName := getEnv("DB_NAME", "kindredcard")
	port := getEnv("PORT", "8080")

	baseURL := getEnv("BASE_URL", fmt.Sprintf("http://localhost:%s", port))

	if u, err := url.Parse(baseURL); err == nil {
		docs.SwaggerInfo.Host = u.Host
		docs.SwaggerInfo.Schemes = []string{u.Scheme}
	}

	// Initialize database
	database, err := db.New(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		logger.Fatal("[APP] Failed to connect to database: %v", err)
	}
	defer database.Close()

	logger.Info("[APP] Connected to database successfully")

	// Initialize handlers
	handler, err := handlers.NewHandler(database, "web/templates", baseURL, ReleaseVersion)
	if err != nil {
		logger.Fatal("[APP] Failed to initialize handlers: %v", err)
	}

	// Initialize CardDAV server
	cardDAVServer := carddav.NewServer(database, !enableTwoWayCardDAV)

	// Initialize Scheduler Service
	schedulerService := scheduler.NewScheduler(database, baseURL)
	schedulerService.Start()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("[APP] Shutting down gracefully...")
		schedulerService.Stop()
		os.Exit(0)
	}()

	// Setup router
	r := mux.NewRouter()

	// Logging middleware
	r.Use(middleware.LoggingMiddleware)

	// Setup check middleware (redirects to /setup if not complete)
	r.Use(middleware.SetupCheckMiddleware(database))

	// Public routes (no auth required)
	r.HandleFunc("/setup", handler.ShowSetup).Methods("GET")
	r.HandleFunc("/setup", handler.ProcessSetup).Methods("POST")
	r.HandleFunc("/login", handler.ShowLogin).Methods("GET")
	r.HandleFunc("/login", handler.ProcessLogin).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Protected web routes
	web := r.PathPrefix("/").Subrouter()
	web.Use(middleware.AuthMiddleware(database))
	web.HandleFunc("/", handler.Index).Methods("GET")
	web.HandleFunc("/events", handler.ShowEvents).Methods("GET")
	web.HandleFunc("/contacts/{id:[0-9]+}", handler.ShowContact).Methods("GET")
	web.HandleFunc("/settings", handler.ShowSettings).Methods("GET")
	web.HandleFunc("/logout", handler.Logout).Methods("GET")

	// Settings: Security
	web.HandleFunc("/settings/password", handler.RotatePassword).Methods("POST")
	web.HandleFunc("/settings/delete-account", handler.DeleteAccount).Methods("DELETE")

	// Utilities
	web.HandleFunc("/utilities/gender-assignment", handler.GenderAssignmentPage).Methods("GET")

	// Protected API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.APIAuthMiddleware(database))
	api.HandleFunc("/contacts", handler.ListContactsAPI).Methods("GET")
	api.HandleFunc("/contacts", handler.CreateContactAPI).Methods("POST")
	api.HandleFunc("/contacts/{id:[0-9]+}", handler.GetContactAPI).Methods("GET")
	api.HandleFunc("/contacts/{id:[0-9]+}", handler.UpdateContactAPI).Methods("PUT")
	api.HandleFunc("/contacts/{id:[0-9]+}", handler.DeleteContactAPI).Methods("DELETE")
	api.HandleFunc("/contacts/{id:[0-9]+}/avatar", handler.UploadAvatarAPI).Methods("POST")
	api.HandleFunc("/contacts/{id:[0-9]+}/avatar", handler.DeleteAvatarAPI).Methods("DELETE")
	api.HandleFunc("/contacts/search", handler.SearchContactsAPI).Methods("GET")
	api.HandleFunc("/contacts/{id:[0-9]+}", handler.PatchContactAPI).Methods("PATCH")

	// Preferences / Theme
	api.HandleFunc("/user/preferences", handler.UpdatePreferencesAPI).Methods("PUT")

	// Settings: Contact Management
	api.HandleFunc("/contacts", handler.DeleteAllContactsAPI).Methods("DELETE")
	api.HandleFunc("/contacts/duplicates", handler.FindDuplicatesAPI).Methods("GET")

	//Settings: Sessions
	api.HandleFunc("/sessions/{id:[0-9]+}", handler.DeleteUserSessionAPI).Methods("DELETE")
	api.HandleFunc("/sessions/revoke-others", handler.DeleteAllOtherUserSessionsAPI).Methods("POST")

	// Upcoming Events API routes
	api.HandleFunc("/events/upcoming", handler.GetUpcomingEventsAPI).Methods("GET")
	api.HandleFunc("/events/count", handler.GetUpcomingEventsCountAPI).Methods("GET")
	api.HandleFunc("/events/today", handler.GetTodaysEventsAPI).Methods("GET")

	// Notifications
	api.HandleFunc("/notification-settings", handler.ListNotificationSettingsAPI).Methods("GET")
	api.HandleFunc("/notification-settings", handler.CreateNotificationSettingAPI).Methods("POST")
	api.HandleFunc("/notification-settings/{id:[0-9]+}", handler.GetNotificationSettingAPI).Methods("GET")
	api.HandleFunc("/notification-settings/{id:[0-9]+}", handler.UpdateNotificationSettingAPI).Methods("PUT")
	api.HandleFunc("/notification-settings/{id:[0-9]+}", handler.DeleteNotificationSettingAPI).Methods("DELETE")
	api.HandleFunc("/notification-settings/{id:[0-9]+}/test", handler.TestNotificationAPI).Methods("POST")

	// Export/Import routes
	api.HandleFunc("/contacts/{id:[0-9]+}/vcard", handler.ExportContactVCardAPI).Methods("GET")
	api.HandleFunc("/contacts/export/vcard", handler.ExportAllVCardsAPI).Methods("GET")
	api.HandleFunc("/contacts/export/json", handler.ExportAllJSONAPI).Methods("GET")
	api.HandleFunc("/contacts/import", handler.ImportVCardsAPI).Methods("POST")

	// Relationship routes
	api.HandleFunc("/relationship-types", handler.GetRelationshipTypesAPI).Methods("GET")
	api.HandleFunc("/contacts/{id:[0-9]+}/relationships", handler.AddRelationshipAPI).Methods("POST")
	api.HandleFunc("/relationships/{rel_id:[0-9]+}", handler.RemoveRelationshipAPI).Methods("DELETE")
	api.HandleFunc("/other-relationships/{rel_id:[0-9]+}", handler.RemoveOtherRelationshipAPI).Methods("DELETE")
	//api.HandleFunc("/relationship-types", handler.CreateRelationshipTypeAPI).Methods("POST")

	// API Tokens
	api.HandleFunc("/tokens", handler.CreateAPIToken).Methods("POST")
	api.HandleFunc("/tokens", handler.ListAPITokens).Methods("GET")
	api.HandleFunc("/tokens/validate", handler.TestAPIToken).Methods("GET")
	api.HandleFunc("/tokens/{id:[0-9]+}", handler.GetAPIToken).Methods("GET")
	api.HandleFunc("/tokens/{id:[0-9]+}", handler.DeleteAPIToken).Methods("DELETE")
	api.HandleFunc("/tokens/{id:[0-9]+}/revoke", handler.RevokeAPIToken).Methods("POST")

	// CardDAV routes (Basic Auth)
	carddav := r.PathPrefix("/carddav").Subrouter()
	carddav.Use(middleware.CardDAVAuthMiddleware(database))
	carddav.PathPrefix("/").Handler(cardDAVServer)

	// Web route for vCard download (needs to be authenticated via cookie)
	r.HandleFunc("/contacts/{id:[0-9]+}/vcard", handler.ExportContactVCardAPI).Methods("GET")

	var customSwaggerJS = `
    const initBranding = () => {
        const topbar = document.querySelector('.topbar-wrapper');
        const mainTopbar = document.querySelector('.topbar');

        if (topbar && mainTopbar) {
            // 1. Make the Topbar Sticky/Frozen
            mainTopbar.style.position = 'sticky';
            mainTopbar.style.top = '0';
            mainTopbar.style.zIndex = '1000';

            // 2. Create the Back Button
            if (!document.getElementById('custom-back-btn')) {
                const backBtn = document.createElement('a');
                backBtn.id = 'custom-back-btn';
                backBtn.innerHTML = '← Back to KindredCard';
                backBtn.href = '/'; 
                backBtn.style.cssText = ` + "`" + `
                    color: white; 
                    margin-right: auto; 
                    text-decoration: none; 
                    font-weight: 600; 
                    font-size: 14px;
                    border: 1px solid rgba(255,255,255,0.3); 
                    padding: 6px 12px; 
                    border-radius: 6px;
                    transition: all 0.2s;
                    display: flex;
                    align-items: center;
                ` + "`" + `;
                
                // Hover effect
                backBtn.onmouseover = () => { backBtn.style.backgroundColor = 'rgba(255,255,255,0.1)'; };
                backBtn.onmouseout = () => { backBtn.style.backgroundColor = 'transparent'; };

                // Insert at the very beginning of the topbar
                topbar.insertBefore(backBtn, topbar.firstChild);

                // Adjust the logo margin so it doesn't look crowded
                const logo = document.querySelector('.link');
                if (logo) logo.style.marginLeft = '20px';
            }
        }
    };

    // Run immediately and also on a slight delay to catch late renders
    initBranding();
    setTimeout(initBranding, 500);
    setTimeout(initBranding, 1500);`

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Points to the generated doc
		httpSwagger.AfterScript(customSwaggerJS),
	))

	// Start server
	logger.Info("[APP] Starting KindredCard server on port %s", port)
	logger.Info("[APP] Web interface: http://localhost:%s", port)
	logger.Info("[APP] API endpoint: http://localhost:%s/api/v1", port)
	logger.Info("[APP] CardDAV endpoint: http://localhost:%s/carddav/", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("[APP] Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
