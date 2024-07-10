# OPDS Proxy

OPDS Proxy provides a minimal web interface over XML-based OPDS feeds. 
Your eReader likely does not support OPDS, but it does have a rudimentary web browser. 
By running your own OPDS Proxy you can allow eReaders to navigate and download books from your library and any OPDS feed without installing custom eReader software.

<p align="center">
    <img src=".github/screenshot.png">
</p>

## Features
- Minimal web interface that works on any web browser
- Multiple OPDS feeds
- Automatically converts your `.epub` files into the proprietary format your eReader requires.
    - Kobo: `*.epub` to `*.kepub` (see [benefits](https://www.reddit.com/r/kobo/comments/vz3nx6/kepub_vs_epub/))
    - Kindle:  `*.epub` to `*.mobi`
    - Other: `*.epub`
- Allows accessing HTTP basic auth OPDS feeds from primitive eReader browsers that don't natively support basic auth. 

## Getting Started
1. Download the latest release binary or pull the latest docker image.
2. Configure your OPDS feeds via environment variables or config file (YML or JSON)
3. Navigate your library and download books to your eReader

## Motivation
KOReader is great and I've been running it for years on my jailbroken Kindle and Kobo eReaders.
To me, the standout feature is its ability to speak OPDS which allows myself / family / friends to have access to my library from anywhere.
That being said, the Kindle / Kobo native reader software is faster, better looking, has a lower learning curve, and doesn't require a complicated installation process that oftens breaks on device updates.

## My Setup
- All books are stored and managed via Calibre (Docker) in standard `.epub` format.
- Calibre's OPDS feed is turned on but not exposed to the outside world.
- OPDS Proxy is pointed to the Calibre OPDS feed.
- eReaders access OPDS Proxy via web browser. Book files are automatically converted to device-specific proprietary format on download.

### Why Not Use An Exisitng Solution?
- Calibre
  - Good for metadata management but not for the web interface.
  - Doesn't support Kobo `.kepub` files natively without installing multiple plugins.
  - HTTP basic auth doesn't work on eReader web browsers.
  - Requires converting and storing eReader specific proprietary formats.
  To me, `.epub` is the [master](https://mixbutton.com/mastering-articles/what-is-the-master-recording/) from which all other copies should be derived.
- Calibre-Web
  - Takes over your entire library and makes it no longer compatible to be managed via Calibre.
  - Dated interface
  - Does too much
  - No automatic conversions
- COPS
  - No longer maintained
  - Dated interface
  - No automatic conversions
- send.djazz.se/
  - Requires uploading private content to unknown server
  - Only supports 1 book at a time
  - Doesn't connect to your existing library 
  - Single User
    - Can't share library with friends / family