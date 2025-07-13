package util

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"sync"
	"timeMonitorClient/global"
)

var (
	// PowerShell process and related variables
	psCmd         *exec.Cmd
	psStdin       io.WriteCloser
	psStdout      io.ReadCloser
	scanner       *bufio.Scanner
	psInitialized bool
	psMutex       sync.Mutex
)

// Initialize the PowerShell process
func initPowerShell() error {
	psMutex.Lock()
	defer psMutex.Unlock()

	if psInitialized {
		return nil
	}

	// Start a persistent PowerShell process with the -Command parameter to run in non-interactive mode
	psCmd = exec.Command("powershell.exe", "-NoProfile", "-NoLogo", "-WindowStyle", "Hidden", "-Command", "-")

	// Get stdin and stdout pipes
	var err error
	psStdin, err = psCmd.StdinPipe()
	if err != nil {
		return err
	}

	psStdout, err = psCmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Start the scanner to read output
	scanner = bufio.NewScanner(psStdout)

	// Start the PowerShell process
	if err := psCmd.Start(); err != nil {
		return err
	}

	// Set up UTF-8 encoding and other initializations
	initScript := `
# Set console output encoding to UTF-8 directly and early
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# Set output encoding for cmdlets like Write-Host
$OutputEncoding = [System.Text.Encoding]::UTF8
# Set console codepage to UTF-8 and suppress its output
chcp 65001 > $null

# Define a special function to output window information
function Get-WindowInfo {
    param()
    
    # Check if the User32 type already exists before adding it
    if (-not ([System.Management.Automation.PSTypeName]'User32').Type) {
        Add-Type -TypeDefinition @"
            using System;
            using System.Runtime.InteropServices;
            using System.Text;

            public class User32 {
                [DllImport("user32.dll")]
                public static extern IntPtr GetForegroundWindow();

                [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
                public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);

                [DllImport("user32.dll")]
                public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
            }
"@
    }
    
    $hwnd = [User32]::GetForegroundWindow()

    if ($hwnd -ne [System.IntPtr]::Zero) {
        # Get Window Title
        $sb = New-Object System.Text.StringBuilder(256)
        [User32]::GetWindowText($hwnd, $sb, $sb.Capacity) | Out-Null
        $windowTitle = $sb.ToString()
        
        # Get Process Name
        $processId = 0
        # Use GetWindowThreadProcessId to get the process ID
        [User32]::GetWindowThreadProcessId($hwnd, [ref]$processId) | Out-Null

        try {
            # Get the process object by ID
            $process = Get-Process -Id $processId -ErrorAction Stop
            $processName = $process.ProcessName
        }
        catch {
            # Handle cases where the process cannot be found
            $processName = "other"
        }
        
        # Output with special prefix for easy parsing
        Write-Output "__WINDOW_TITLE__:$windowTitle"
        Write-Output "__PROCESS_NAME__:$processName"
    } else {
        Write-Output "__WINDOW_TITLE__:No foreground window"
        Write-Output "__PROCESS_NAME__:other"
    }
}

Write-Output "__INIT_COMPLETE__"
`
	// Send initialization script
	if _, err := psStdin.Write([]byte(initScript + "\n")); err != nil {
		return err
	}

	// Wait for initialization confirmation
	for scanner.Scan() {
		line := scanner.Text()
		if line == "__INIT_COMPLETE__" {
			break
		}
	}

	psInitialized = true
	return nil
}

// Send a command to PowerShell and return the output with markers
func executeMarkedCommand(command string) ([]string, error) {
	// Create a unique command with clear start/end markers
	markedCommand := `
Write-Output "__CMD_START__"
` + command + `
Write-Output "__CMD_END__"
`

	// Send the command
	if _, err := psStdin.Write([]byte(markedCommand + "\n")); err != nil {
		return nil, err
	}

	// Collect output between markers
	var output []string
	collecting := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "__CMD_START__" {
			collecting = true
			continue
		}

		if line == "__CMD_END__" {
			break
		}

		if collecting {
			output = append(output, line)
		}
	}

	return output, nil
}

// Clean up PowerShell resources
func CleanupPowerShell() {
	if psInitialized {
		psMutex.Lock()
		defer psMutex.Unlock()

		if psStdin != nil {
			// Try to exit gracefully
			psStdin.Write([]byte("exit\n"))
			psStdin.Close()
		}

		// Force kill if still running
		if psCmd != nil && psCmd.Process != nil {
			psCmd.Process.Kill()
		}

		psInitialized = false
	}
}

func PowerShellOutput() []string {
	// Initialize PowerShell if not already done
	if !psInitialized {
		if err := initPowerShell(); err != nil {
			global.Error("初始化PowerShell失败")
			return []string{"other", "other"}
		}
	}

	psMutex.Lock()
	defer psMutex.Unlock()

	// Execute the window info command
	output, err := executeMarkedCommand("Get-WindowInfo")
	if err != nil {
		global.Error("执行PowerShell命令失败")
		return []string{"other", "other"}
	}

	// Parse the prefixed output
	var windowTitle, processName string

	for _, line := range output {
		if strings.HasPrefix(line, "__WINDOW_TITLE__:") {
			windowTitle = strings.TrimPrefix(line, "__WINDOW_TITLE__:")
		} else if strings.HasPrefix(line, "__PROCESS_NAME__:") {
			processName = strings.TrimPrefix(line, "__PROCESS_NAME__:")
		}
	}

	if windowTitle == "" {
		windowTitle = "other"
	}

	if processName == "" {
		processName = "other"
	}

	return []string{windowTitle, processName}
}
