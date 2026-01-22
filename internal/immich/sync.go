package immich

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

// SyncService handles syncing between Immich and KindredCard
type SyncService struct {
	client *Client
	db     *db.Database
	userID int
}

// NewSyncService creates a new sync service
func NewSyncService(client *Client, database *db.Database, userID int) *SyncService {
	return &SyncService{
		client: client,
		db:     database,
		userID: userID,
	}
}

// Match represents a potential match between Immich person and KindredCard contact
type Match struct {
	ImmichPerson Person
	Contact      *models.Contact
	MatchType    string // "exact", "nickname", "linked"
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	TotalPeople        int
	MatchedContacts    int
	AvatarsSynced      int
	BirthdaysExtracted int
	Errors             []string
	Matches            []Match
}

// findBestMatch finds the best matching contact for an Immich person
func (s *SyncService) findBestMatch(person Person, contacts []*models.Contact) Match {
	var bestMatch Match
	bestMatch.ImmichPerson = person

	personNameLower := strings.ToLower(person.Name)

	for _, contact := range contacts {

		// Check exact name match
		if strings.ToLower(contact.FullName) == personNameLower {
			return Match{
				ImmichPerson: person,
				Contact:      contact,
				MatchType:    "exact",
			}
		}

		// Check nickname match
		if contact.Nickname != "" && strings.ToLower(contact.Nickname) == personNameLower {
			return Match{
				ImmichPerson: person,
				Contact:      contact,
				MatchType:    "nickname",
			}
		}

		// Check given name (firstname) match
		if contact.GivenName != "" && strings.ToLower(contact.GivenName) == personNameLower {
			return Match{
				ImmichPerson: person,
				Contact:      contact,
				MatchType:    "given",
			}
		}

		// Regex: First Initial + Last Name
		if contact.FamilyName != "" {
			var firstInitials []string
			if contact.GivenName != "" {
				firstInitials = append(firstInitials, strings.ToLower(string(contact.GivenName[0])))
			}
			if contact.Nickname != "" {
				firstInitials = append(firstInitials, strings.ToLower(string(contact.Nickname[0])))
			}

			if len(firstInitials) > 0 {
				// Joins initials into (a|b)
				initialPattern := strings.Join(firstInitials, "|")

				// QuoteMeta ensures a last name like "St. John" doesn't treat the "." as a regex wildcard
				familyNameEscaped := regexp.QuoteMeta(strings.ToLower(contact.FamilyName))

				// Pattern: Starts with initial, followed by anything, ending with family name
				// Example: ^(a|b).*doe$
				pattern := fmt.Sprintf("^(%s).*%s$", initialPattern, familyNameEscaped)

				matched, _ := regexp.MatchString(pattern, personNameLower)
				if matched {
					return Match{
						ImmichPerson: person,
						Contact:      contact,
						MatchType:    "regex",
					}
				}
			}
		}
	}

	bestMatch.MatchType = "none"
	return bestMatch
}

// GetPotentialMatches returns potential matches for review
func (s *SyncService) GetPotentialMatches() ([]Match, error) {
	logger.Debug("[IMMICH] Getting potential matches")

	// Get existing links from DB
	alreadyLinkedMap, _ := s.db.GetLinkedImmichIDs(s.userID)

	// Get all people from Immich
	allPeople, err := s.client.GetAllPeople()
	if err != nil {
		return nil, fmt.Errorf("failed to get people: %w", err)
	}

	// Filter people: Only keep those NOT in the map
	var availablePeople []Person
	for _, p := range allPeople {
		if !alreadyLinkedMap[p.ID] {
			availablePeople = append(availablePeople, p)
		}
	}

	// Get all contacts without an immich link
	contacts, err := s.db.GetUnlinkedImmichContacts(s.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	// Match against the available pool
	var matches []Match
	for _, person := range availablePeople {
		match := s.findBestMatch(person, contacts)
		if match.Contact != nil {
			matches = append(matches, match)
		}
	}

	logger.Info("[IMMICH] Found %d potential matches", len(matches))
	return matches, nil
}

// GetPotentialMatches returns potential matches for review
func (s *SyncService) GetAllLinkedContacts() ([]Match, error) {
	logger.Debug("[IMMICH] Getting linked contacts")

	contacts, err := s.db.GetContactsByURL(s.userID, s.client.BaseURL, "immich")
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	matches := make([]Match, 0, len(contacts))
	for _, contact := range contacts {

		if len(contact.URLs) == 0 {
			continue
		}

		personID := utils.ExtractIDFromImmichURL(contact.URLs[0].URL)

		immichPersonPTR, _ := s.client.GetPerson(personID)

		matches = append(matches, Match{
			Contact:      contact,
			ImmichPerson: *immichPersonPTR,
			MatchType:    "linked",
		})
	}

	logger.Info("[IMMICH] Found %d linked matches", len(matches))
	return matches, nil
}
