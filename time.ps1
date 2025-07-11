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
    Write-Host "title: $windowTitle"

    $processId = 0

    try {
        $process = Get-Process -Id $processId -ErrorAction Stop
        Write-Host "process: $($process.ProcessName)"
    }
    catch {
        Write-Host "error process (ID: $processId)"
    }
} else {
    Write-Host "Error: No foreground window found."
}