Unicode true
!include "MUI2.nsh"
!include "LogicLib.nsh"

Name "PC Remote v5.0"
OutFile "PCRemoteRustSetup.exe"
InstallDir "$PROGRAMFILES64\PCRemote"
RequestExecutionLevel admin
SetCompressor /SOLID lzma

!define MUI_ABORTWARNING
!define MUI_ICON "..\..\server-rust\assets\pc_remote.ico"
!define MUI_UNICON "..\..\server-rust\assets\pc_remote.ico"
!define MUI_FINISHPAGE_RUN "$INSTDIR\pcremote-server.exe"
!define MUI_FINISHPAGE_RUN_TEXT "Launch PC Remote"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "PC Remote (Required)" Main
    SectionIn RO
    SetShellVarContext current
    SetOutPath "$INSTDIR"

    ; Cleanly stop running PC Remote & legacy Go v4 processes before overwriting
    nsExec::ExecToStack 'taskkill /F /IM pcremote-server.exe /T'
    nsExec::ExecToStack 'taskkill /F /IM PCRemoteDashboard.exe /T'
    nsExec::ExecToStack 'taskkill /F /IM cloudflared.exe /T'
    Sleep 1000

    ; Stop and remove legacy Go NSSM service if present
    nsExec::ExecToStack 'nssm stop PCRemoteServer'
    nsExec::ExecToStack 'nssm remove PCRemoteServer confirm'
    nsExec::ExecToStack '"$INSTDIR\nssm.exe" stop PCRemoteServer'
    nsExec::ExecToStack '"$INSTDIR\nssm.exe" remove PCRemoteServer confirm'

    ; Remove obsolete v4.x files from in-place installation directory
    Delete "$INSTDIR\PCRemoteDashboard.exe"
    Delete "$INSTDIR\cloudflared.exe"
    Delete "$INSTDIR\ScreenOff.exe"
    Delete "$INSTDIR\.env"
    Delete "$INSTDIR\nssm.exe"
    Delete "$INSTDIR\favicon.ico"

    ; Remove legacy v4.x & preview shortcuts / startup entries
    Delete "$SMSTARTUP\PCRemoteServer.lnk"
    Delete "$SMSTARTUP\PC Remote Rust.lnk"
    Delete "$DESKTOP\PCRemote Dashboard.lnk"
    Delete "$SMPROGRAMS\PC Remote Server.lnk"
    Delete "$SMPROGRAMS\PCRemote\PCRemote Dashboard.lnk"
    RMDir "$SMPROGRAMS\PCRemote"
    Delete "$SMPROGRAMS\PC Remote Rust\PC Remote.lnk"
    Delete "$SMPROGRAMS\PC Remote Rust\Setup guide.lnk"
    Delete "$SMPROGRAMS\PC Remote Rust\Uninstall.lnk"
    RMDir "$SMPROGRAMS\PC Remote Rust"
    DeleteRegValue HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "PCRemoteServer"
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "PCRemoteServer"

    ; Remove duplicate %LOCALAPPDATA%\PCRemoteRust binaries while preserving data directory
    Delete "$LOCALAPPDATA\PCRemoteRust\pcremote-server.exe"
    Delete "$LOCALAPPDATA\PCRemoteRust\README.md"
    Delete "$LOCALAPPDATA\PCRemoteRust\configure-firewall.ps1"
    Delete "$LOCALAPPDATA\PCRemoteRust\Uninstall.exe"

    SetOverwrite on
    ClearErrors
    File "..\..\server-rust\target\release\pcremote-server.exe"
    ${If} ${Errors}
        Abort "Could not write pcremote-server.exe. Close PC Remote and retry the installation."
    ${EndIf}
    File "README.md"
    File "..\..\scripts\rust\configure-firewall.ps1"
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    ; Start Menu & Desktop shortcuts
    CreateDirectory "$SMPROGRAMS\PC Remote"
    CreateShortcut "$SMPROGRAMS\PC Remote\PC Remote.lnk" "$INSTDIR\pcremote-server.exe"
    CreateShortcut "$SMPROGRAMS\PC Remote\Setup guide.lnk" "$INSTDIR\README.md"
    CreateShortcut "$SMPROGRAMS\PC Remote\Configure Firewall.lnk" "powershell.exe" '-ExecutionPolicy Bypass -File "$INSTDIR\configure-firewall.ps1" -Lan -Tailscale'
    CreateShortcut "$SMPROGRAMS\PC Remote\Uninstall.lnk" "$INSTDIR\Uninstall.exe"
    CreateShortcut "$DESKTOP\PC Remote.lnk" "$INSTDIR\pcremote-server.exe"

    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "DisplayName" "PC Remote v5.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "DisplayVersion" "5.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "DisplayIcon" "$INSTDIR\pcremote-server.exe,0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "UninstallString" '$\"$INSTDIR\Uninstall.exe$\"'
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "InstallLocation" "$INSTDIR"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote" "NoRepair" 1
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemoteRust"
SectionEnd

Section /o "Start after Windows sign-in" Startup
    SetShellVarContext current
    CreateShortcut "$SMSTARTUP\PC Remote.lnk" "$INSTDIR\pcremote-server.exe" "--background"
SectionEnd

Section "Uninstall"
    SetShellVarContext current

    ; Stop running server cleanly before removing binary
    nsExec::ExecToStack 'taskkill /F /IM pcremote-server.exe /T'
    Sleep 500

    ClearErrors
    Delete "$INSTDIR\pcremote-server.exe"
    ${If} ${Errors}
        Abort "Close PC Remote from the System Tray and retry uninstalling."
    ${EndIf}

    Delete "$SMSTARTUP\PC Remote.lnk"
    Delete "$SMSTARTUP\PC Remote Rust.lnk"
    Delete "$DESKTOP\PC Remote.lnk"
    Delete "$SMPROGRAMS\PC Remote\PC Remote.lnk"
    Delete "$SMPROGRAMS\PC Remote\Setup guide.lnk"
    Delete "$SMPROGRAMS\PC Remote\Configure Firewall.lnk"
    Delete "$SMPROGRAMS\PC Remote\Uninstall.lnk"
    RMDir "$SMPROGRAMS\PC Remote"

    Delete "$INSTDIR\README.md"
    Delete "$INSTDIR\configure-firewall.ps1"
    Delete "$INSTDIR\Uninstall.exe"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"

    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "PCRemoteRust"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemote"
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\PCRemoteRust"
    ; Pairing credentials/data in %LOCALAPPDATA%\PCRemoteRust\data are retained for reinstall. Tailscale is untouched.
SectionEnd
