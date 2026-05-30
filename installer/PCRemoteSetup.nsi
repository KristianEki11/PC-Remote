!include "MUI2.nsh"
!include "nsDialogs.nsh"
!include "LogicLib.nsh"
!include "FileFunc.nsh"

Name "PC Remote Controller v2.2.5"
OutFile "PCRemoteSetup.exe"
InstallDir "$PROGRAMFILES64\PCRemote"
RequestExecutionLevel admin

; Variables
Var Dialog
Var Label
Var TextPIN
Var PIN

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "license.txt"
Page custom CreatePINPage ValidatePINPage
!insertmacro MUI_PAGE_INSTFILES

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Function CreatePINPage
    !insertmacro MUI_HEADER_TEXT "Set your PIN" "Choose a PIN for the PC Remote Android App."

    nsDialogs::Create 1018
    Pop $Dialog

    ${If} $Dialog == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 12u "Set your PIN (numbers only, min 4 digits):"
    Pop $Label

    ${NSD_CreateText} 0 15u 100% 12u ""
    Pop $TextPIN

    nsDialogs::Show
FunctionEnd

Function ValidatePINPage
    ${NSD_GetText} $TextPIN $PIN

    StrLen $0 $PIN
    ${If} $0 < 4
    ${OrIf} $0 > 8
        MessageBox MB_OK|MB_ICONEXCLAMATION "PIN must be between 4 and 8 digits."
        Abort
    ${EndIf}

    ; Validate only digits
    StrCpy $1 0
    Loop:
        IntCmp $1 $0 Done
        StrCpy $2 $PIN 1 $1
        ${If} $2 < "0"
        ${OrIf} $2 > "9"
            MessageBox MB_OK|MB_ICONEXCLAMATION "PIN must contain only numbers."
            Abort
        ${EndIf}
        IntOp $1 $1 + 1
        Goto Loop
    Done:
FunctionEnd

Section "MainSection" SEC01
    SetOutPath "$INSTDIR"

    ; Kill any running server processes first to unlock files
    nsExec::ExecToStack 'taskkill /F /IM pcremote-server.exe'

    ; Cleanup any old NSSM service if it exists from previous installations
    nsExec::ExecToStack 'nssm stop PCRemoteServer'
    nsExec::ExecToStack 'nssm remove PCRemoteServer confirm'
    nsExec::ExecToStack '"$INSTDIR\nssm.exe" stop PCRemoteServer'
    nsExec::ExecToStack '"$INSTDIR\nssm.exe" remove PCRemoteServer confirm'

    ; Install files
    File "..\server\dist\pcremote-server.exe"
    File "..\server\dist\sendkey.exe"
    File "..\server\favicon.ico"

    ; Create logs directory
    CreateDirectory "$INSTDIR\logs"

    ; Write .env file
    FileOpen $0 "$INSTDIR\.env" w
    FileWrite $0 "APP_PIN=$PIN$\r$\n"
    FileWrite $0 "PORT=8000$\r$\n"
    FileClose $0

    ; Register as Startup Application in HKLM Registry Run key
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "PCRemoteServer" '"$INSTDIR\pcremote-server.exe"'

    ; Open firewall
    nsExec::ExecToLog 'netsh advfirewall firewall add rule name="PCRemote Server" dir=in action=allow protocol=TCP localport=8000'

    ; Start the application directly in the user session
    Exec '"$INSTDIR\pcremote-server.exe"'

    MessageBox MB_OK|MB_ICONINFORMATION "PC Remote Server installed successfully!$\r$\nServer is running on port 8000.$\r$\nConnect your Android app to: http://[your-pc-ip]:8000"

    MessageBox MB_YESNO|MB_ICONEXCLAMATION "⚠️ IMPORTANT: Network Profile Check$\r$\n$\r$\nMake sure your WiFi is set to PRIVATE network.$\r$\n$\r$\nOpen: Settings -> Network -> [Your WiFi] -> Properties$\r$\nSet to: Private Network$\r$\n$\r$\nWithout this, your phone cannot connect even if$\r$\nthe server is running correctly.$\r$\n$\r$\nOpen Network Settings now?" IDYES OpenNetwork IDNO SkipNetwork

OpenNetwork:
    ExecShell "open" "ms-settings:network"
SkipNetwork:

    ; Write uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

Section "Uninstall"
    ; Kill running instance
    nsExec::ExecToStack 'taskkill /F /IM pcremote-server.exe'

    ; Stop and remove service (cleanup old installations if they exist)
    nsExec::ExecToStack 'nssm stop PCRemoteServer'
    nsExec::ExecToStack 'nssm remove PCRemoteServer confirm'

    ; Remove registry startup
    DeleteRegValue HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "PCRemoteServer"

    ; Remove firewall rule
    nsExec::ExecToLog 'netsh advfirewall firewall delete rule name="PCRemote Server"'

    ; Delete files
    Delete "$INSTDIR\pcremote-server.exe"
    Delete "$INSTDIR\sendkey.exe"
    Delete "$INSTDIR\favicon.ico"
    Delete "$INSTDIR\.env"
    Delete "$INSTDIR\uninstall.exe"
    
    ; Delete logs folder if user confirms
    MessageBox MB_YESNO|MB_ICONQUESTION "Do you want to delete the logs folder?" IDNO KeepLogs
    RMDir /r "$INSTDIR\logs"
KeepLogs:

    RMDir "$INSTDIR"
SectionEnd
