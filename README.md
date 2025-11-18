  # Disgo CLI &middot; [![discord](https://img.shields.io/discord/1297292231299956788?color=5865F2&logo=discord&logoColor=white)](https://discord.com/invite/VzF9UFn2aB) [![ci](https://github.com/involvex/disgo-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/involvex/disgo-cli/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/involvex/disgo-cli)](https://goreportcard.com/report/github.com/involvex/disgo-cli) [![license](https://img.shields.io/github/license/involvex/disgo-cli?logo=github)](https://github.com/involvex/disgo-cli/blob/main/LICENSE) [![npm](https://img.shields.io/npm/v/@involvex/disgo-cli)](https://www.npmjs.com/package/@involvex/disgo-cli)

Disgo CLI is a lightweight, secure, and feature-rich Discord terminal client. Features web-based login interface with QR code authentication.

![Preview](.github/preview.png)

## Features

- Lightweight
- Configurable
- Mouse & clipboard support
- Attachments
- Notifications
- 2-Factor & QR code authentication
- Discord-flavored markdown

## Installation

### NPM Package (Recommended)

Install globally via npm:

```bash
npm install -g @involvex/disgo-cli
```

### Prebuilt binaries

You can download and install a [prebuilt binary here](https://github.com/involvex/disgo-cli/releases) for Windows, macOS, or Linux.

### Package managers

- Arch Linux: `yay -S disgo-cli-git` (coming soon)
- Gentoo (available on the guru repos as a live ebuild): `emerge net-im/disgo-cli` (coming soon)
- FreeBSD: `pkg install disgo-cli` (coming soon)
- Nix (NixOS, home-manager)
  - Upstream flake installation: Add `inputs.disgo-cli.url = "github:involvex/disgo-cli"`. Install using `inputs.disgo-cli.homeModules.default` (`.enable, .package, .settings TOML`).
- Windows (Scoop): Coming soon

### Building from source

```bash
git clone https://github.com/involvex/disgo-cli
cd disgo-cli
go build .
```

### Wayland clipboard support

`x11-dev` is required for X11 clipboard compatibility:

- Ubuntu: `apt install xwayland`
- Arch Linux: `pacman -S xorg-xwayland`

## Usage

### Basic Usage

1. Run the `disgo-cli` executable with no arguments:

```bash
disgo-cli
```

2. For web-based login, use the `--serve` flag:

```bash
disgo-cli --serve
```

This starts a web server at `http://localhost:8080` with a Discord-themed interface for QR code or API-based login.

> If you are logging in using an authentication token, provide the `token` command-line flag (eg: `--token "OTI2MDU5NTQxNDE2Nzc5ODA2.Yc2KKA.2iZ-5JxgxG-9Ub8GHzBSn-NJjNg"`). Alternatively, set the value of the `DISGO_CLI_TOKEN` environment variable to the authentication token. The token is stored securely in the default OS-specific keyring.

### NPM Scripts

If installed via npm, you can use these commands:

```bash
# Start the application
npm start

# Start with web login interface
npm run serve

# Run in development mode
npm run dev

# Build the binary
npm run build

# Clean build artifacts
npm run clean
```

## Configuration

The configuration file allows you to configure and customize the behavior, keybindings, and theme of the application.

- Unix: `$XDG_CONFIG_HOME/disgo-cli/config.toml` or `$HOME/.config/disgo-cli/config.toml`
- Darwin: `$HOME/Library/Application Support/disgo-cli/config.toml`
- Windows: `%AppData%/disgo-cli/config.toml`

Disgo CLI uses the default configuration if a configuration file is not found in the aforementioned path; however, the default configuration file is not written to the path. [The default configuration can be found here](./internal/config/config.toml).

## FAQ

### Manually adding token to keyring

Do this if you get the error:

> failed to get token from keyring: secret not found in keyring

#### Windows

Run the following command in a terminal window. Replace `YOUR_DISCORD_TOKEN` with your authentication token.

```sh
cmdkey /add:disgo-cli /user:token /pass:YOUR_DISCORD_TOKEN
```

#### MacOS

Run the following command in a terminal window. Replace `YOUR_DISCORD_TOKEN` with your authentication token.

```sh
security add-generic-password -s disgo-cli -a token -w "YOUR_DISCORD_TOKEN"
```

#### Linux

1. Start the keyring daemon.

```sh
eval $(gnome-keyring-daemon --start)
export $(gnome-keyring-daemon --start)
```

2. Create the `login` keyring if it does not exist already. See [GNOME/Keyring](https://wiki.archlinux.org/title/GNOME/Keyring) for more information.

3. Run the following command to create the `token` entry.

```sh
secret-tool store --label="Discord Token" service disgo-cli username token
```

4. When it prompts for the password, paste your token, and hit enter to confirm.

> [!IMPORTANT]
> Automated user accounts or "self-bots" are against Discord's Terms of Service. I am not responsible for any loss caused by using "self-bots" or Disgo CLI.
