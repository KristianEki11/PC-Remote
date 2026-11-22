# PC Remote Rust setup

This installer is per Windows user and installs separately from the legacy Go service. Default ports: HTTPS 8443 for the phone, loopback-only HTTP 8444 for the local setup page. Starting PC Remote opens the setup page; logon startup uses `--background` to suppress the browser.

The current server runs in the signed-in user's desktop session. It is **not a Windows service** and is not reachable before sign-in. "Start after Windows sign-in" is an optional installer component. Stop PC Remote using its local setup dashboard before installing an update or uninstalling; the installer does not kill processes. Existing Go installations and their services are not modified. Do not remove the Go installation until the features you use have passed acceptance testing.

## First setup

1. Install PC Remote Rust and open PC Remote from the Start menu.
2. For LAN-only operation, connect the phone to the same network. Allow the installed executable through Windows Firewall only on the intended private network/interface; the installer does not broadly open firewall ports. Guest Wi-Fi/client isolation may prevent direct access.
3. For internet access, install official [Tailscale](https://tailscale.com/download) on the PC and phone. Sign in both devices to **your own tailnet**. Other PC Remote users use their own accounts; nobody must join the developer's network.
4. If Tailscale is already installed, retain the existing profile, exit-node choice, routes and SSH policy. Ensure the phone is permitted to reach this PC's application port under that tailnet's access policy. Organization-managed tailnets may require administrator approval.
5. Open the local setup page, generate a pairing QR, and scan it with PC Remote on the phone. Pairing is required even for other members of the same tailnet. Keep the QR private and pair again if its one-time token expires.
6. On Windows, Tailscale's **Run Unattended** setting keeps its network connection active after logout/restart. This does not make the PC Remote user-session process start before sign-in. Device key expiry is separate: review it in Tailscale's admin console if unattended access is important. Do not disable SSH check-mode or change an organization's policy just to use PC Remote.

Android permits only one active VPN at a time. Another VPN can conflict with Tailscale; PC Remote does not turn other VPNs off. Tailscale's free Personal plan is subject to its noncommercial-use terms; business users must review their applicable plan.

## Windows Firewall

On first use Windows may ask whether the app can receive connections. Allow the private network you intend to use. If policy blocks the prompt, an administrator can add a narrowly scoped inbound TCP rule for the installed executable and port 8443. Tailscale access policy and Windows Firewall both need to permit traffic; enabling one does not bypass the other. Never expose the local setup port 8444 or create a router port-forward.

The installed `configure-firewall.ps1 -Lan -Tailscale -WhatIf` previews explicit rules in elevated PowerShell; the source copy is under `scripts/rust/`. Remove `-WhatIf` to apply them. Pass `-Binary` if you chose a custom installation folder. The LAN rule permits only local-subnet traffic on Private/Domain profiles; the Tailscale rule is restricted to the detected Tailscale adapter and its IPv4 address range. The script stops if the adapter is ambiguous or rules already exist. It does not alter Tailscale policy, VPN routes, SSH, DNS, or an existing firewall rule. Organization-managed PCs may require the administrator to apply the equivalent policy. Uninstall does not remove manually added administrator firewall rules; an administrator can remove the named `PCRemoteRust-LAN-8443` and `PCRemoteRust-Tailscale-8443` rules if no longer needed.

## Data, updates and rollback

The default data directory is `%LOCALAPPDATA%\PCRemoteRust\data`. It contains device identity and pairing state: do not upload it with bug reports. Uninstall preserves this directory. Reinstall under the same Windows account to retain its data protection context. The `PCREMOTE_DATA_DIR` environment variable overrides it; `PCREMOTE_PORT` overrides HTTPS port 8443.

To roll back, close Rust PC Remote and remove its optional startup shortcut, then use the existing Go installation. A phone may need to pair again with the selected server identity. Do not copy Rust credentials into Go or vice versa. There is no automatic service replacement or credential conversion in this installer.

## Troubleshooting

- If the PC is powered off, asleep, or no user has signed in, remote control cannot be assumed available.
- If LAN works but internet does not, check that both Tailscale clients are connected to the intended tailnet, then check the port policy and firewall. A timeout alone cannot identify which component blocked traffic.
- If Tailscale says authentication is required, authenticate in Tailscale. PC Remote cannot override expired/revoked network identity.
- If a peer can reach the server but pairing is revoked, pair explicitly again. Tailscale network membership is not application authorization.

References: [Run Unattended](https://tailscale.com/docs/how-to/run-unattended), [key expiry](https://tailscale.com/docs/features/access-control/key-expiry), [VPN conflicts](https://tailscale.com/docs/reference/faq/other-vpns), [pricing](https://tailscale.com/pricing).
