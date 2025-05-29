# Night Owls Go — PWA Front‑end Structure (Static Build)

A revised blueprint for a **pure Svelte 5** front‑end compiled to a static bundle and served by the existing Go binary—no second back‑end layer required.

---

## 1 Objectives

* **Volunteer reminders** – Web‑Push notifications for shift reminders, attendance prompts, and incident updates.
* **Admin oversight** – CRUD dashboards for schedules, bookings, and incident reports.
* **Offline resilience** – Core views cached; queued actions sync when back online.
* **Mobile‑first installability** – PWA manifest, service‑worker asset caching, and home‑screen install support.

---

## 2 Features & Responsibilities

### 2.1 Volunteer Experience

#### Dashboard

* **Upcoming Shifts** – list of the volunteer’s booked slots.
* **Quick Book** – one‑click booking panel showing the next seven days’ vacancies.
* **Recent Reports** – last incident reports filed by the volunteer.
* **Notifications** – latest system messages and reminders.

#### Shift Management

* **View Available Shifts** – calendar or list of open slots.
* **Book Shifts** – select and confirm a slot.
* **Buddy System** – add a buddy (registered or ad‑hoc) at booking time.
* **Cancel Booking** – withdraw from a slot (subject to cut‑off rules).
* **Mark Attendance** – confirm presence once patrol is complete.

#### Reporting

* **Submit Incident Reports** – file shift outcomes.
* **Report History** – browse previous submissions.
* **Severity Levels** – tag 0–2 seriousness.

#### Profile & Settings

* **Notification Preferences** – opt‑in/out per channel.
* **Shift History** – archive of attended patrols.
* **Offline Mode** – toggle data to keep cached.

### 2.2 Administrator Tools

#### Programme Management

* **Schedule Management** – create, edit, or retire recurring shift schedules.
* **Shift Calendar** – global view of booked vs vacant slots.
* **Volunteer Directory** – manage accounts and phone numbers.
* **Reports Dashboard** – triage and follow‑up on incidents.
* **Statistics** – participation metrics and export.

#### Volunteer Management

* **Account Verification** – approve new sign‑ups via OTP.
* **Assign Roles** – promote/demote administrators.
* **Activity Tracking** – monitor attendance trends.
* **Announcements** – push one‑off messages to all or subsets.

#### Report Oversight

* **Review & Filter** – search, sort, and mark reports as resolved.

### 2.3 PWA Capabilities

#### Offline Support

* View booked shifts offline.
* Browse recently cached available shifts.
* Queue attendance marks while offline.
* Draft and queue reports for later sync.

#### Push Notifications

* Shift reminders (24 h and 1 h before start).
* Booking confirmations and cancellations.
* Schedule changes impacting a user’s bookings.
* Admin announcements.

---

## 3 Why *not* add another back‑end?

| Extra back‑end use‑case                    | Relevance here                                | Verdict |
| ------------------------------------------ | --------------------------------------------- | ------- |
| **Server‑Side Rendering / SEO**            | Minimal: app is auth‑gated & primarily mobile | ✖ Skip  |
| **Edge‑side personalisation / AB testing** | Out of scope for volunteer tool               | ✖ Skip  |
| **GraphQL/BFF aggregation layer**          | Go API already concise; no 3rd‑party mash‑ups | ✖ Skip  |
| **Security through proxy (hide API)**      | Same origin already; CORS not an issue        | ✖ Skip  |
| **Asset streaming / ISR**                  | Static data, no SEO need                      | ✖ Skip  |

> **Conclusion:** A static bundle hosted by the Go server keeps the deployment simple, reduces infra surface area, and still leverages service‑worker + TanStack Query for dynamic behaviour.

---

## 4 Chosen Stack

| Concern            | Technology                                                      |
| ------------------ | --------------------------------------------------------------- |
| Framework / build  | **Svelte 5 + Vite** (`@sveltejs/vite-plugin-svelte`)            |
| Styling            | **Tailwind CSS v4** + **shadcn‑svelte @next** component library |
| Data‑fetch / cache | **TanStack Query** via `@tanstack/svelte-query`                 |
| Tables             | **TanStack Table** via `@tanstack/svelte-table`                 |
| Icons / UI bits    | lucide‑svelte                                                   |
| Testing            | Vitest + Playwright                                             |

---

## 5 Directory Layout

```
frontend/
├─ index.html                  # entry point, injects PWA meta
├─ vite.config.ts              # SW, Tailwind, shadcn plugin wiring
├─ src/
│  ├─ main.ts                 # bootstrap Svelte app, QueryClient init
│  ├─ app.css                 # Tailwind base + shadcn layers
│  ├─ lib/
│  │  ├─ api/
│  │  │  ├─ client.ts        # fetch wrapper (JWT inject)
│  │  │  ├─ queries/         # TanStack Query hooks
│  │  │  └─ mutations/
│  │  ├─ components/
│  │  │  ├─ ui/              # shadcn‑svelte re‑exports if customised
│  │  │  ├─ tables/
│  │  │  └─ forms/
│  │  ├─ hooks/
│  │  │  └─ usePush.ts       # subscription logic
│  │  ├─ stores/
│  │  │  ├─ auth.ts          # user + JWT
│  │  │  └─ settings.ts
│  │  └─ pwa/
│  │     ├─ service-worker.ts
│  │     ├─ manifest.webmanifest
│  │     └─ push.ts
│  ├─ pages/                  # simple page components for routing
│  │  ├─ Router.svelte        # tiny Svelte SPA router (e.g., svelte‑navaid)
│  │  ├─ AuthLogin.svelte
│  │  ├─ Dashboard.svelte
│  │  ├─ ShiftsAvailable.svelte
│  │  ├─ ShiftsMine.svelte
│  │  └─ admin/
│  │     ├─ Schedules.svelte
│  │     ├─ Bookings.svelte
│  │     └─ Reports.svelte
│  └─ static/                 # PWA icons
└─ tests/
   ├─ unit/
   └─ e2e/
```

### Bundling & Serve

* `vite build` emits to `frontend/dist/`.
* **Go server** embeds `dist` via `//go:embed` or falls back to `http.FileServer`.
* SPA router falls back to `index.html` for unknown paths.

---

## 6 Push‑Notification Flow

1. **Subscribe** in `usePush.ts` → send to `/push/subscribe`.
2. Backend worker sends Web‑Push via VAPID.
3. SW displays notification; on click opens relevant route.
4. Logout triggers unsubscribe.

---

## 7 State & Offline Strategy

* TanStack Query handles caching & background refetch.
* Failed mutations queued via Background Sync API in SW.
* Auth store persists token in `IndexedDB` (via `idb-keyval`).

---

## 8 UI Guidelines (Tailwind + shadcn‑svelte)

* Use shadcn primitives (`Button`, `Input`, `Dialog`) with Tailwind utility overrides.
* Colour palette:

  * `primary` — #0F172A (midnight)
  * `accent`  — #EAB308 (amber‑500) for shift status highlights.
* All components support dark mode by default (`dark:` classes).

---

## 9 Build, Lint & CI

* `eslint`, `prettier`, `stylelint`, `svelte-check`.
* Lighthouse PWA CI on PRs via GitHub / GitLab CI.
* Release artefact: Go binary + embedded `dist` → single container image.

---

## 10 Progressive Roadmap

A staged checklist so the team can layer functionality incrementally and avoid scope overload. Tackle one milestone at a time; ship, test, and iterate before moving on.

---

### Milestone 0 – Scaffold & CI

* [ ] Initialise repo with *Svelte 5 + TypeScript* template.
* [ ] Integrate Tailwind CSS v4 and **shadcn‑svelte @next**.
* [ ] Configure Vite, ESLint, Prettier, Stylelint, Vitest.
* [ ] Add CI pipeline: build, unit tests, lint.

### Milestone 1 – Auth & API Client

* [ ] Create `lib/api/client.ts` with JWT injection & automatic refresh.
* [ ] Build **Login** (phone) and **OTP Verify** pages.
* [ ] Persist JWT in IndexedDB; protect routes via store guard.

### Milestone 2 – Volunteer Core V1

* [ ] **Dashboard** with *Upcoming Shifts* (read‑only).
* [ ] **View Available Shifts** list with date filter.
* [ ] **Book Shifts** (no buddy yet) & **Cancel Booking**.
* [ ] Display in‑app **Notifications** list.

### Milestone 3 – Reporting & Buddy Support

* [ ] Add *Buddy* field to booking flow (registered or ad‑hoc).
* [ ] **Submit Incident Reports** form.
* [ ] **Report History** page.

### Milestone 4 – PWA Layer (Offline + Push)

* [ ] Add service‑worker with asset precache & runtime cache for core GET routes.
* [ ] Queue failed mutations (attendance, reports) via Background Sync.
* [ ] Implement Web‑Push subscription workflow (`usePush.ts`).
* [ ] Receive push notifications; test 24 h & 1 h shift reminders.

### Milestone 5 – Admin Tools MVP

* [ ] **Schedule Management** CRUD screens.
* [ ] **Shift Calendar** overview (read‑only grid).
* [ ] **Volunteer Directory** with role toggle.

### Milestone 6 – Advanced Admin & Analytics

* [ ] **Reports Dashboard** with filters & status update.
* [ ] **Statistics** charts (participation, incident frequency).
* [ ] **Announcements** broadcast modal & push integration.

### Milestone 7 – Polish, QA & Release

* [ ] Lighthouse PWA audit ≥ 90 on Performance, Accessibility, Best Practices.
* [ ] End‑to‑end Playwright tests for critical flows.
* [ ] UX feedback cycle & micro‑interaction polish.
* [ ] Build final container image: Go API + embedded `dist`.

---

> Check off each box as you ship. The roadmap is intentionally linear—feel free to parallelise tasks if capacity allows, but keep milestones in order to preserve a deployable product at every stage.
