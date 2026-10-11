# Staffing plan

- Profil: **LEAN** | Misi: **ENHANCE** | Archetype: **desktop-windows, android-app**
- Flags: desktop, installable, mobile, ui
- Cluster: dasar 3 | FE 3 | BE 3 | App 3
- **Total spawn: 33** (Board 4 + Head 8 + Worker 20 + Staff 1)
- MAX_TOTAL_SPAWNS disarankan: **37** (termasuk buffer respawn) | MAX_CONCURRENT: **6**

| Divisi | Peran | ID | Jenis | Jml | Tier | Workspace | Playbook |
|---|---|---|---|---|---|---|---|
| Board | ATLAS (Chair) | `board-chair` | board | 1 | pro | inherit | board-council |
| Board | Contrarian | `board-contrarian` | board | 1 | pro | inherit | board-council |
| Board | Executor | `board-executor` | board | 1 | pro | inherit | board-council |
| Board | Outsider | `board-outsider` | board | 1 | pro | inherit | board-council |
| Strategy & Research | Head of Strategy & Research (CPO+CRO) | `head-strategy` | head | 1 | pro | inherit | product-management+research-intel |
| Strategy & Product | Spec Writer | `spec-writer` | worker | 1 | flash | inherit | product-management |
| Research & Intelligence | Explorer (read-only) | `explorer` | worker | 2 | flash | inherit | research-intel |
| Architecture & Decisions | Chief Architect | `chief-architect` | head | 1 | pro | inherit | architecture |
| Design | Head of Design (UX+UI) | `head-design` | head | 1 | pro | inherit | ux-design+ui-design+design-system |
| Design · UX | UX Designer (IA, flow, wireframe) | `ux-designer` | worker | 1 | pro | inherit | ux-design+accessibility-i18n |
| Design · UI | UI Designer (visual, hi-fi prototype) | `ui-designer` | worker | 1 | pro | inherit | ui-design+design-system |
| Design | Design Critic (UX+UI review) | `design-critic` | worker | 1 | pro | inherit | ux-design+ui-design |
| Engineering · Mobile | VP Mobile (Android) | `vp-mobile` | head | 1 | pro | inherit | mobile-android |
| Engineering · Desktop | VP Desktop (Windows) | `vp-desktop` | head | 1 | pro | inherit | desktop-windows |
| Engineering · Mobile | Builder Android | `builder-android` | worker | 3 | pro x2 + flash x1 | branch | mobile-android+design-system+accessibility-i18n |
| Engineering · Desktop | Builder Desktop | `builder-desktop` | worker | 3 | pro x2 + flash x1 | branch | desktop-windows+design-system+accessibility-i18n |
| Platform & DevOps | VP Platform & DevOps | `vp-platform` | head | 1 | pro | branch | devops-platform |
| QA & Verification | VP QA & Verification | `vp-qa` | head | 1 | pro | inherit | qa-testing+bug-response+performance-reliability |
| QA & Verification | Tester (independen) | `tester` | worker | 2 | flash | branch | qa-testing+accessibility-i18n+performance-reliability |
| QA & Verification | Critic (code review) | `critic-code` | worker | 1 | pro | inherit | qa-testing |
| Security & Compliance | Security Auditor (AppSec, dependency, secrets) | `security-auditor` | worker | 1 | pro | inherit | security-appsec+pentest-sandbox+privacy-compliance |
| Bug Response | SWAT (Investigator/Fixer) | `swat` | worker | 2 | pro | branch | bug-response |
| Release & Docs | Release Manager (Integration) | `release-manager` | head | 1 | pro | branch | release-integration |
| Release & Docs | Docs Writer | `docs-writer` | worker | 1 | flash | branch | docs-writing |
| Operations & Facilities | Janitor (Office Boy) | `janitor` | staff | 1 | flash | inherit | back-office |
| Internal Audit | Auditor (verifikasi ulang bukti) | `auditor` | worker | 1 | pro | inherit | qa-testing |
