# KindredCard - Personal CRM & CardDAV Server

Your personal CRM system built in Go with PostgreSQL backend and CardDAV server capabilities. Manage your contacts through a simple web interface and sync them with any CardDAV-compatible client (Apple Contacts, Android, Thunderbird, etc.).

## Screenshots

<img src="https://github.com/steveredden/KindredCard/wiki/assets/images/index.png"/>
<p align="center">
  <img src="https://github.com/steveredden/KindredCard/wiki/assets/images/index-modal1.png" width=49%/><img src="https://github.com/steveredden/KindredCard/wiki/assets/images/index-modal2.png" width=49%/>
</p>

## Features

- ✅ **Full Contact Management**: Names, emails, phones, addresses, organizations, notes, gender, and more
- ✅ **Contact Utilities**: Easily transform Phone Numbers, assign contacts Genders, and more - [link](https://github.com/steveredden/KindredCard/wiki/Utilities)
- ✅ **vCard Import/Export**: Import or Export your .vcf (vCard) files - link(coming)
- ✅ **Events Dashboard**: Keep track of your contacts' important life events - [link](https://github.com/steveredden/KindredCard/wiki/Events)
- ✅ **CardDAV Server**: Sync contacts with any CardDAV-compatible client - [link](https://github.com/steveredden/KindredCard/wiki/CardDAV)
- ✅ **RESTful API**: OpenAPI 3.0 specification for programmatic access - [link](https://github.com/steveredden/KindredCard/wiki/REST-API)
- ✅ **SMTP Notifications**: Bring your own SMTP server for event digests delivered to your inbox - [link](https://github.com/steveredden/KindredCard/wiki/Notifications#smtp-prerequisites)
- ✅ **Discord Notifications**: Webhook integration to notify when events are ocurring - [link](https://github.com/steveredden/KindredCard/wiki/Notifications#discord-prerequisites)

## Quick Start with Docker

Check out the [Wiki](https://github.com/steveredden/KindredCard/wiki/Docker) for explicit instructions

## Acknowledgments

- Built with [go-vcard](https://github.com/emersion/go-vcard) for vCard parsing
- Inspired by [monicahq/monica](https://github.com/monicahq/monica/tree/4.x)
- CardDAV implementation based on [RFC 6350](https://datatracker.ietf.org/doc/html/rfc6350) and [RFC 6352](https://datatracker.ietf.org/doc/html/rfc6352)
- Definitely Viiiiibe-Coding involved

## Support

For issues, questions, or suggestions, please open an issue on GitHub.

## Contributing

Please review the [CONTRIBUTING.md](CONRTIBUTING.md)