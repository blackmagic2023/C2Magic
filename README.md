# C2Magic

C2Magic is a lightweight Command and Control (C2) server designed for managing multiple client devices through a centralized server interface. C2Magic allows users to send commands to specific devices or broadcast commands to all connected devices, with a unique identifier system for easy management. The server can handle multiple client connections concurrently and displays output in a user-friendly, color-coded format. **C2Magic was designed to be ran on Kali Linux 2024.2+, Debian 12 BookWorm!**

## Updates
- **Adding additional payloads**
- **Adding Additional Features for Automation**
- **Adding Client Features for Stealth and Evasion**

## Features

- **Unique Device ID System**: Assigns a unique ID to each device upon connection, making it easier to select and manage devices.
- **Concurrent Command Handling**: Allows multiple client devices to connect and process commands concurrently.
- **Customizable Command Broadcast**: Send commands to a specific device by its unique ID or broadcast to all connected devices.
- **Command Output in Color**: The output from each device is displayed in green for easy differentiation.
- **Intuitive Command Interface**: Keeps users in the command session for a device, with the ability to send multiple commands without returning to the main menu.
- **Client Stub Builder**: Allows you to compile a stub.exe with a provided undetectable reverse powershell payload.

## Installation

1. **Clone the Repository**
```bash
git clone https://github.com/blackmagic2023/C2Magic.git
cd C2Magic
```
2. Install Dependencies C2Magic requires Go and the fatih/color package for colorized output, install it by running:
```bash
go mod init FaithColor
go get github.com/fatih/color@latest
go mod tidy
```
3. Build the Server
```bash
go build C2Magic.go
```

## Usage

To start the C2Magic server, specify the port number:
```bash
./C2Magic <Port>
```

## Server Command Menu

The main command menu has the following options:

1. Send Command to All Clients: Broadcast a command to all connected clients.
2. Send Command to a Specific Client: Send a command to a specific client by entering its unique device ID.
3. List Connected Clients: Display all connected clients, including their ID and IP address.
4. Create Client Stub: Compiles a reverse powershell payload to connect back to your server into executable format.

## Sending Commands to Devices

Once connected to a device, you can send multiple commands without returning to the main menu. Type `back` to return to the main menu.

## Contributions

Contributions are welcome! If you’d like to contribute to C2Magic, please fork the repository and create a pull request with your modifications.

## Disclaimer

This project is intended for educational and ethical testing purposes only. Use responsibly and ensure you have permission before connecting or managing devices through C2Magic.

