
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Threading;

public class WinMedia {
    [DllImport("user32.dll")]
    public static extern void keybd_event(byte bVk, byte bScan, uint dwFlags, uint dwExtraInfo);

    public const byte VK_SHIFT = 0x10;
    public const byte VK_N = 0x4E;
    public const byte VK_P = 0x50;
    public const byte VK_K = 0x4B;
    public const byte VK_J = 0x4A;
    public const byte VK_L = 0x4C;
    public const byte VK_MEDIA_NEXT = 0xB0;
    public const byte VK_MEDIA_PREV = 0xB1;
    public const byte VK_MEDIA_PLAY_PAUSE = 0xB3;
    public const uint KEYEVENTF_KEYUP = 0x0002;
    public const uint KEYEVENTF_EXTENDEDKEY = 0x0001;

    public static void SendNext() {
        // 1. Send hardware media next key
        keybd_event(VK_MEDIA_NEXT, 0, KEYEVENTF_EXTENDEDKEY, 0);
        keybd_event(VK_MEDIA_NEXT, 0, KEYEVENTF_EXTENDEDKEY | KEYEVENTF_KEYUP, 0);
        Thread.Sleep(30);

        // 2. Send YouTube/Chromium next video shortcut (Shift + N)
        keybd_event(VK_SHIFT, 0, 0, 0);
        keybd_event(VK_N, 0, 0, 0);
        keybd_event(VK_N, 0, KEYEVENTF_KEYUP, 0);
        keybd_event(VK_SHIFT, 0, KEYEVENTF_KEYUP, 0);
    }

    public static void SendPrev() {
        // 1. Send hardware media prev key
        keybd_event(VK_MEDIA_PREV, 0, KEYEVENTF_EXTENDEDKEY, 0);
        keybd_event(VK_MEDIA_PREV, 0, KEYEVENTF_EXTENDEDKEY | KEYEVENTF_KEYUP, 0);
        Thread.Sleep(30);

        // 2. Send YouTube/Chromium prev video shortcut (Shift + P)
        keybd_event(VK_SHIFT, 0, 0, 0);
        keybd_event(VK_P, 0, 0, 0);
        keybd_event(VK_P, 0, KEYEVENTF_KEYUP, 0);
        keybd_event(VK_SHIFT, 0, KEYEVENTF_KEYUP, 0);
    }

    public static void SendPlayPause() {
        keybd_event(VK_MEDIA_PLAY_PAUSE, 0, KEYEVENTF_EXTENDEDKEY, 0);
        keybd_event(VK_MEDIA_PLAY_PAUSE, 0, KEYEVENTF_EXTENDEDKEY | KEYEVENTF_KEYUP, 0);
    }

    public static void SendSeekForward() {
        // 'L' for YouTube 10s skip
        keybd_event(VK_L, 0, 0, 0);
        keybd_event(VK_L, 0, KEYEVENTF_KEYUP, 0);
    }

    public static void SendSeekBackward() {
        // 'J' for YouTube 10s rewind
        keybd_event(VK_J, 0, 0, 0);
        keybd_event(VK_J, 0, KEYEVENTF_KEYUP, 0);
    }
}
"@

# 1. Try Windows WinRT SMTC session first
$smtcHandled = $false
try {
    [void][System.Reflection.Assembly]::LoadWithPartialName("System.Runtime.WindowsRuntime")
    [void][Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]

    $asTaskMethod = [System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Length -eq 1 } | Select-Object -First 1
    $genericMethod = $asTaskMethod.MakeGenericMethod([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager])
    $task = $genericMethod.Invoke($null, @([Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()))
    $manager = $task.Result

    $session = $manager.GetCurrentSession()
    if ($session) {
        $info = $session.GetPlaybackInfo()
        $act = "play_pause"
        if ($act -eq "play_pause" -and $info.Controls.IsPlayPauseToggleEnabled) {
            $asyncOp = $session.TryTogglePlayPauseAsync()
            $genericMethodBool = $asTaskMethod.MakeGenericMethod([bool])
            $taskBool = $genericMethodBool.Invoke($null, @($asyncOp))
            $taskBool.Wait()
            if ($taskBool.Result) { $smtcHandled = $true }
        } elseif ($act -eq "next" -and $info.Controls.IsNextEnabled) {
            $asyncOp = $session.TrySkipNextAsync()
            $genericMethodBool = $asTaskMethod.MakeGenericMethod([bool])
            $taskBool = $genericMethodBool.Invoke($null, @($asyncOp))
            $taskBool.Wait()
            if ($taskBool.Result) { $smtcHandled = $true }
        } elseif ($act -eq "prev" -and $info.Controls.IsPreviousEnabled) {
            $asyncOp = $session.TrySkipPreviousAsync()
            $genericMethodBool = $asTaskMethod.MakeGenericMethod([bool])
            $taskBool = $genericMethodBool.Invoke($null, @($asyncOp))
            $taskBool.Wait()
            if ($taskBool.Result) { $smtcHandled = $true }
        }
    }
} catch {}

# 2. Always execute targeted hardware & browser keyboard emission if SMTC did not fully handle it
if (-not $smtcHandled) {
    switch ("play_pause") {
        "play_pause"     { [WinMedia]::SendPlayPause() }
        "next"           { [WinMedia]::SendNext() }
        "prev"           { [WinMedia]::SendPrev() }
        "seek_forward"   { [WinMedia]::SendSeekForward() }
        "seek_backward"  { [WinMedia]::SendSeekBackward() }
    }
}
