<h1 align="center">
  <a href="https://github.com/immanent-tech/go-syndication">
    <!-- Please provide path to your logo here -->
    <!-- <img src="docs/images/logo.svg" alt="Logo" width="100" height="100"> -->
  </a>
</h1>

<div align="center">
  go-syndication
  <br />
  <br />
  <a href="https://github.com/immanent-tech/go-syndication/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+">Report a Bug</a>
  ·
  <a href="https://github.com/immanent-tech/go-syndication/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+">Request a Feature</a>
  .
  <a href="https://github.com/immanent-tech/go-syndication/issues/new?assignees=&labels=question&template=04_SUPPORT_QUESTION.md&title=support%3A+">Ask a Question</a>
</div>

<div align="center">
<br />

[![Project license](https://img.shields.io/github/license/immanent-tech/go-syndication.svg?style=flat-square)](LICENSE)

[![Pull Requests welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/immanent-tech/go-syndication/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
[![code with love by joshuar](https://img.shields.io/badge/%3C%2F%3E%20with%20%E2%99%A5%20by-joshuar-ff1414.svg?style=flat-square)](https://github.com/joshuar)

</div>

<details open="open">
<summary>Table of Contents</summary>

- [About](#about)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Installation](#installation)
- [Usage](#usage)
  - [Encoding and Decoding](#encoding-and-decoding)
  - [Validation](#validation)
  - [Generic Feed/Item Types](#generic-feeditem-types)
  - [Command Line Interface (CLI)](#command-line-interface-cli)
- [Roadmap](#roadmap)
- [Support](#support)
- [Project Assistance](#project-assistance)
- [Contributing](#contributing)
- [Authors \& Contributors](#authors--contributors)
- [Security](#security)
- [License](#license)

</details>

---

## About

`go-syndication` is Go package for dealing with various feed syndication formats. It supports:

- RSS
- Atom
- JSONFeed
- OPML
- Various RSS/Atom extensions such as media, Dublin Core, iTunes, and GooglePlay, with more to come...

The package can read and write all formats. It includes built-in validation of elements.

### Built With

- [openapi-codegen](https://github.com/oapi-codegen/oapi-codegen/).
- [validator](https://github.com/go-playground/validator).
- [bluemonday](https://github.com/microcosm-cc/bluemonday).

## Getting Started

### Installation

```shell
go get github.com/immanent-tech/go-syndication
```

## Usage

### Encoding and Decoding

You can decode a `io.Reader` containing feed data into a format using `func Decode[T any](namespace string, rd
io.Reader) (T, error)`. `T` would be one of `*atom.Feed`, `*rss.RSS` or `*jsonfeed.Feed`:

```go
// Decode RSS feed data.
// use bytes.NewReader(data) for a []byte
rss, err := Decode[*rss.RSS]("", data)
```

Likewise, `func Encode[T any](feed T) ([]byte, error)` can be used to encode feed data:

```go
data, err := Encode[*rss.RSS](rss)
```

### Validation

By default, encoding/decoding performs no validation. As long as the XML data is well-formed and can be read by the Go
XML parser, your feed data should be encoded/decoded. This means the package will handle invalid feed data (feeds that
don't adhere to the RSS/Atom specs).

If you want to check that the feed data is valid, you can use the validation sub-package:

```go
// Decode atom feed data.
// use bytes.NewReader(data) for a []byte
rss, err := Decode[*rss.RSS]("", data)

// Validate the atom feed.
err := validation.ValidateStruct(rss)
```

### Generic Feed/Item Types

In addition to providing the source-specific `atom.Feed`, `rss.RSS` and `jsonfeed.Feed` types and their item
counterparts, this library provides generic `feeds.Feed` and `feeds.Item` types, that wrap the source types with common
methods for accessing their fields.

Use `func NewDecoder[T any](data io.Reader) (*Feed, error)` to read data into the generic object:

```go
// data is a []byte containing an atom feed.
feed, err = feeds.NewDecoder[*rss.RSS](bytes.NewReader(data))
```

`Feed` exposes the original data as the `FeedSource`, which can be converted back to the original source format with a
type conversion:

```go
atom, ok := feed.FeedSource.(*atom.Feed)
```

This gives you the best of both worlds; a generic container with common methods for canonical fields across all formats,
with access to the original source to manipulate the format directly as needed.

### Command Line Interface (CLI)

A basic CLI can be found in `cmd/` that can be used for basic reading/writing of feeds using the library.

To fetch and display feed data from a URL:

```shell
go run github.com/immanent-tech/go-syndication/cmd@latest fetch http://my.site/feed
```

To read a file containing feed data:

```shell
go run github.com/immanent-tech/go-syndication/cmd@latest parse /path/to/my/feed.xml
```

The commands will auto-detect a supported feed format.

## Roadmap

See the [open issues](https://github.com/immanent-tech/go-syndication/issues) for a list of proposed features (and known issues).

- [Top Feature Requests](https://github.com/immanent-tech/go-syndication/issues?q=label%3Aenhancement+is%3Aopen+sort%3Areactions-%2B1-desc) (Add your votes using the 👍 reaction)
- [Top Bugs](https://github.com/immanent-tech/go-syndication/issues?q=is%3Aissue+is%3Aopen+label%3Abug+sort%3Areactions-%2B1-desc) (Add your votes using the 👍 reaction)
- [Newest Bugs](https://github.com/immanent-tech/go-syndication/issues?q=is%3Aopen+is%3Aissue+label%3Abug)

## Support

Reach out to the maintainer at one of the following places:

- [GitHub issues](https://github.com/immanent-tech/go-syndication/issues/new?assignees=&labels=question&template=04_SUPPORT_QUESTION.md&title=support%3A+)
- Contact options listed on [this GitHub profile](https://github.com/joshuar)

## Project Assistance

If you want to say **thank you** or/and support active development of go-syndication:

- Add a [GitHub Star](https://github.com/immanent-tech/go-syndication) to the project.
- Tweet about the go-syndication.
- Write interesting articles about the project on [Dev.to](https://dev.to/), [Medium](https://medium.com/) or your
  personal blog.

Together, we can make go-syndication **better**!

## Contributing

First off, thanks for taking the time to contribute! Contributions are what make the open-source community such an
amazing place to learn, inspire, and create. Any contributions you make will benefit everybody else and are **greatly
appreciated**.

Please read [our contribution guidelines](CONTRIBUTING.md), and thank you for being involved!

## Authors & Contributors

The original setup of this repository is by [joshuar](https://github.com/joshuar).

For a full list of all authors and contributors, see [the contributors
page](https://github.com/immanent-tech/go-syndication/contributors).

## Security

go-syndication follows good practices of security, but 100% security cannot be assured.
go-syndication is provided **"as is"** without any **warranty**. Use at your own risk.

_For more information and to report security issues, please refer to our [security documentation](SECURITY.md)._

## License

This project is licensed under the **MIT license**.

See [LICENSE](LICENSE) for more information.
