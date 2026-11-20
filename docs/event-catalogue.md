# Event catalogue

Draft of the events and fields we need so the product questions in
[product-questions.md](./product-questions.md) can be answered. One row is one
thing we record. The last column lists every question that row helps with
(one event can serve many questions).

The full set of questions is larger than we want to implement in one go. We
split it so the first ship can answer whether people arrive, explore, return,
become owners, share on the Viewer, and actually draw. Now also covers basic
command usage and help interactions (which commands and variants run, key
versus strip, written help and video opens, including before or after first
use). Deeper help and command diagnosis — play-to-end, invalid selections,
failure reasons, settings — wait until that pipe is trusted.

- **Now** — the first slice we will record and answer, including that basic
  command and help coverage.
- **Later** — still wanted, but only after the Now pipe is trusted. Later adds
  events and fields (and richer parameters on a few Now events, such as command
  detail); it does not replace Now.

The matching Now/Later cut for what we want to answer is in
[product-questions.md](./product-questions.md). Names and parameters are not
final until we say we are willing to implement the Now list. Each section has
one table with the same three columns. An event’s **full parameter list for
its tier lives in the section where it first appears**; later sections say
“as in …” and add only questions.

The common envelope (type, UTC time, browser UUID, schema version) is defined
in [event-schema.md](./event-schema.md). Every row below assumes that envelope.
The **Parameters** column is only the extras for that event.

## Working notes

- A **visit** is treated as the run of events from the same browser UUID with no
  idle gap longer than **20 minutes**. That threshold is opinionated and can be
  changed later if visits feel too long or too chopped-up. Visit length is
  preferably `Application opened` (or first event) through `Visit ended`; if
  `Visit ended` never arrived, fall back to the last event timestamp.
- A **first visit** is the first visit we ever see for that browser UUID. We do
  not need a separate “is first visit” field on every event if the UUID and
  timestamps are enough to derive it.
- Prefer a few generic event types with parameters over one type per button.
- **`Visit ended`:** fire on the page’s `pagehide` handler, read
  `mutationsSinceSave` then, and send with `navigator.sendBeacon` (or
  `fetch` + `keepalive`) so the request can still leave as the page dies. Not
  perfect (crash, killed process, failed beacon), but the intended signal for
  leave and for “ended dirty.”
- **Desktop newcomer footer:** standard-footer mode only clears when this device
  becomes an owner (save, open an own drawing, or OK on the name dialog for a
  new drawing). There is no separate “dismiss footer” control, so we do **not**
  emit a footer-left event. Matching “left the footer?” questions were dropped
  from [product-questions.md](./product-questions.md) as duplicates of owning
  drawings.
- **Unsaved changes** on `Drawing saved` means the toolbar counter
  `mutationsSinceSave` (what makes Save available and shows “N unsaved
  changes”) at the moment of save. When emitting `Drawing saved`, capture
  `mutationsSinceSave` immediately before the save operation resets the
  counter.
- **Look/select-only visits (Later):** a visit counts as only looking and
  selecting when `mutationsSinceSave` never left 0 in that visit. Do not use
  the value at `Visit ended` alone — a mid-visit Save resets the counter to 0.
  Treat the visit as having changed the drawing if any `Drawing saved` in the
  visit reported a pre-clear count above 0, or `Visit ended` reported a count
  above 0, or (if we add it later) an explicit peak / ever-mutated flag for
  the visit. Undo increments this count; it does not reduce it.
- **Home screen choice versus sheets:** emit on **both** faces when the visitor
  taps a home/landing choice. For “how often did they tap A versus B on Home”,
  count only `Home screen choice`. Opening the examples list or “How DrawExact
  works” still emits `Example list opened` / `Help opened` for the step trail —
  do not add those on top of the Home choice when comparing taps. Desktop plain
  splash example picks stay `Example chosen`, not this event.
- **Help ids** match the controls people see. The orientation sheet is
  `how_drawexact_works` (on-screen “How DrawExact works”). Older copy that said
  “What is DrawExact?” means the same sheet.

## Now

### Canonical event types (Now)

`Application opened`, `Home screen choice`, `Example list opened`,
`Example chosen`, `Help opened` (orientation, move-around, and any command’s
written help), `Video opened` (newcomer zoom and command help videos),
`Create own started`, `Visit ended`, `Connect Drive shown`, `Drive connected`,
`Blank drawing started`, `Drawing saved`, `Own drawing opened`,
`Drawing shared`, `Command executed` (thin: command, variant, key vs strip).

### Visits

Covers **First visit** and **Coming back** in the Now product questions.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Application opened | Which face (`viewer` or `desktop`); how they arrived (`deep_link` or `plain`); drawing name when arrival was a deep link; whether this browser has already connected Google Drive; whether this browser has already opened one of the visitor’s own drawings | How far someone gets through the newcomer steps (this is the start of the trail). Whether first visits arrive on a deep link or the plain address. How long a first visit is (with `Visit ended`, or last event if that never arrived). On a deep link, whether they only stay on that drawing or also open the examples list (needs `Example list opened` too). On desktop plain arrival, whether they pick an example from the entry splash or leave without picking (needs `Example chosen` or its absence). How often someone with a first visit opens again another day. How long the gap is before they come back. How many separate days they come back in the first week or month. How many come back once then disappear versus keep coming. What counts as a “regular” user (an analysis rule on this same data, for example three different days) and how many first-timers reach it. Whether a return arrives on a deep link or the plain address. On mobile, how many returning visitors still see the newcomer start screen (opened-own flag still false). On desktop, how many returning visitors still see the newcomer footer — this device is not yet an owner (derive from any prior `Drawing saved`, `Own drawing opened`, or `Blank drawing started` for that browser; Drive connected alone does not clear the footer). After they have saved or opened an own drawing, how often they keep coming back (with Owning drawings milestones). |
| Visit ended | Which face; `mutationsSinceSave` at leave (toolbar unsaved-changes count; usually 0 on the Viewer). Fired from `pagehide` via sendBeacon / keepalive — see working notes | How long a visit is (end timestamp). How often users leave with unsaved changes, and how many (`mutationsSinceSave` at leave). |
| Home screen choice | Which face; which home form (`newcomer` or `owners`); which choice. Viewer newcomer: `show_drawings`, `how_drawexact_works`, or `already_have_drawings`. Viewer owners: `your_drawings` (large), and quiet `show_drawings` / `how_drawexact_works` when tapped. Desktop owners/comfortable landing: `your_drawings`, `new_drawing`, or `revisit_examples`. This is the tap on Home/landing; sheets that open still emit their own events — for A-versus-B rates on Home, count only this event | How far through the newcomer steps when the trail goes via Home. On the Viewer, how often people tap “Show me some drawings” versus “How DrawExact works”, and “I already have drawings” / “Your drawings”. On desktop, how often returners pick Your drawings, New drawing, or revisit examples from the landing. |
| Example list opened | Which face | How far through the newcomer steps. Whether a deep-link arrival also opens the list of other examples. Which examples get picked (this is the “opened the list” half). How often the list is opened and left with no drawing picked (absence of `Example chosen` after this). On a return visit, whether they mostly look at examples again (with `Example chosen`; draw waits on `Command executed`). |
| Example chosen | Which face; which example (the name shown in the list or splash) | How far through the newcomer steps. On desktop plain arrival, whether they pick an example from the splash. Which examples newcomers pick. How often the list is opened with no pick (this event missing after `Example list opened`). On a return visit, whether they mostly look at examples again. |
| Help opened | Which face; which help (`how_drawexact_works`, `how_to_move_around`, or written help for a named command). `how_to_move_around` and command written help are desktop only — do not emit them on the Viewer. When it is command written help, which command | How far through the newcomer steps. Whether they open “How DrawExact works”. Whether they take “How to move around in this drawing?” on desktop. On a return visit, whether they open help (orientation or a command’s written help). Command written-help rates and before/after use are under Help + Videos. |
| Video opened | Desktop only. Which video; which command when it is a command help video; whether it was the short reminder or the longer guide. Covers the newcomer zoom video and command help videos | How far through the newcomer steps. Whether they open the zoom video that the move-around note points at. Command help-video order, counts, and before-use are under Help + Videos. |
| Create own started | Which face. On the Viewer this is the handoff-to-computer sheet; on desktop this is the create-own path (and may lead to Connect Drive) | How far through the newcomer steps. Whether they try “Create your own drawing” (handoff on the Viewer, create path on desktop). |

### Owning drawings

Covers **Becoming someone who owns drawings** in the Now product questions. Also
supplies milestones for “come back after saving or opening an own drawing.”

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Connect Drive shown | Which face; what they were trying to do when the ask appeared. On desktop the trigger may be create own, open own drawings, or save. On the Viewer the trigger is open own drawings only (not create own — that is handoff, not Drive — and not save) | How often people see the Connect Google Drive ask and leave without connecting (no later `Drive connected` for that browser, or none soon after this show). |
| Drive connected | Which face | How many first-visitors connect Drive (with `Application opened` to know they are first-time). How often the Connect ask is followed by a real connect versus abandoned. |
| Blank drawing started | Desktop only | How many users who connect Google Drive go on to create their first owned drawing by starting blank, and how many choose that route versus saving a copy of an example (with `Drawing saved`). |
| Drawing saved | Desktop only. Whether this was a copy of an example, a first save of a blank new drawing, or a later save of an already-owned drawing; example name when it was a copy of an example; whether this is the browser’s first owned drawing; `mutationsSinceSave` at save time (the toolbar unsaved-changes count) | How many users who connect Google Drive go on to create their first owned drawing, either by starting blank (`Blank drawing started`) or by saving a copy of an example — and how many choose each route. How often someone’s first owned drawing comes from saving an example they were looking at. Come-back-after-own-drawing in Visits. How many unsaved changes people build up before saving. |
| Own drawing opened | Which face; whether this is the first time this browser has opened an own drawing | How many people return on a later visit and open an owned drawing. Come-back-after-own-drawing in Visits. |

### Drawing and commands

Covers **Drawing and commands** in the Now product questions. This section
introduces **no new event types** beyond thin `Command executed` — but it is
worth keeping as its own table so the catalogue lines up with that
product-questions heading.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Command executed | Desktop only; which command; which variant if the command has variants; whether it was fired by keystroke or by the command strip. **Now is thin** — Grow end, paste copy count/angle, and similar are Later | How often each command and each variant is used, and which variants are rare. Whether Undo follows some commands more than others (Undo is this event with command Undo, compared to the previous command in the visit). How often commands are fired by keystroke versus the command strip. On a return visit, whether they actually draw (with examples/help events). |

### Help + Videos

Covers **Help + Videos** in the Now product questions. No new event types —
reuses `Help opened`, `Video opened`, and `Command executed` from Visits /
Drawing and commands so that heading has a matching catalogue table.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Video opened | As in Visits | In what order different commands’ help videos get opened. For a given command, whether the longer guide or the short reminder is opened first. How many times each command help video is opened. Whether people open a command’s videos before using that command (with `Command executed`). |
| Help opened | As in Visits | Whether people open any command’s written help. For a given command, whether they open its help before first use or only after (with `Command executed`). After written help, how often they run that command in the same visit. |
| Command executed | As in Drawing and commands | Before/after and same-visit follow-through for command help videos and written help (with `Video opened` / `Help opened`). |

### Save and Drive

Covers **Save and Drive** in the Now product questions.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Own drawing opened | As in Owning drawings | How often people open an existing drawing of their own, versus opening an example (with `Example chosen` and deep-link example arrivals below). |
| Example chosen | As in Visits | How often people open an example by picking from the list or splash, versus opening an own drawing (with `Own drawing opened`). |
| Application opened | As in Visits | How often people open an example by arriving on a deep link (`deep_link` to an example, not an owned drawing), versus opening an own drawing (with `Own drawing opened`). Count together with `Example chosen` for the own-versus-example rate; a visit that deep-links and later also picks from the list counts as two example opens. |
| Drawing saved | As in Owning drawings | How many unsaved changes people build up before saving. |
| Visit ended | As in Visits | How often users leave with unsaved changes, and how many (`mutationsSinceSave` at leave). |

### Mobile (Viewer)

Covers **Mobile Users** in the Now product questions. Most rows filter existing
events to `face = viewer`.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Application opened | As in Visits (read with `face = viewer`) | How many first mobile visits arrive on a deep link versus the plain address. How far a first mobile visit gets (with the mobile steps below). Of browsers that only ever use mobile (`face = viewer`), how many come back another day. On mobile, how many returning visitors still see the newcomer start screen (same opened-own flag as in Visits). Whether mobile visits stay on examples or move toward share / Connect / open-own over time (with the events below). |
| Home screen choice | As in Visits (`face = viewer`) | How often people tap “Show me some drawings” versus “How DrawExact works” on the newcomer start screen. How often people tap “I already have drawings” or “Your drawings” before Connect. |
| Example list opened | As in Visits (`face = viewer`) | How far through a first Viewer visit (sheet opened — not a substitute for `Home screen choice` tap rates). |
| Example chosen | As in Visits (`face = viewer`) | How far through a first Viewer visit. Whether a share was from a list pick (with `Drawing shared`). |
| Help opened | As in Visits (`face = viewer`) | How far through a first Viewer visit when they open “How DrawExact works” (sheet opened — not a substitute for `Home screen choice` tap rates). |
| Create own started | As in Visits (`face = viewer` — handoff sheet) | How far through a first Viewer visit. How often people tap “Create your own drawing” on the Viewer. |
| Drawing shared | Viewer; purpose (`friend` or `self_handoff`); method (`share_sheet` or `copy`); whether the drawing on screen came from a deep-link arrival or from a list pick | How often people share a drawing. Whether they copy or share-sheet the link that hands off to a computer (`self_handoff`). Whether a share was from a deep-linked example or from something picked in the list. |
| Connect Drive shown | As in Owning drawings (`face = viewer`, trigger open own drawings) | How often Connect appears after “I already have drawings” / “Your drawings” (with `Home screen choice`). |
| Drive connected | As in Owning drawings (`face = viewer`) | Whether Connect on the phone is followed by opening an own drawing or is a bounce (with `Own drawing opened`). |
| Own drawing opened | As in Owning drawings (`face = viewer`) | Bounce after Connect versus open. How often someone opens their own drawing on the Viewer at all. |

### General site info

Covers **General Site Info** in the Now product questions. No new event types —
aggregates over the Now catalogue.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Application opened | As in Visits | How many different browsers visit in a week (distinct browser UUIDs). How long a typical session lasts (with `Visit ended`, or last event if that never arrived). |
| Visit ended | As in Visits | End timestamp for typical session length. |
| *(analysis buckets)* | Not an emit — classify recorded events as first-visit / discovery, help, or real drawing when querying | Of everything we record, how much is first-visit behaviour, how much is help, and how much is real drawing. |
| Command executed | As in Drawing and commands | Which commands are rarely used. |
| Help opened | As in Visits | Which help is rarely opened (orientation, move-around, or a command’s written help). |
| Video opened | As in Visits | Whether the newcomer zoom video or command help videos are rarely opened. Which commands’ videos get opened a lot. |

## Later

### Canonical additions (Later)

New event types: `Video closed`, `Selection options expanded`, `Invalid
selection`, `Command failed`, `Setting changed`, `Key used`, `Export PNG`.

Same events, richer parameters: `Command executed` gains Grow end / target,
paste copy count and angle, and similar.

### Help and videos

Covers **Help + Videos** in the Later product questions (play-to-end and
selection-options help). Command video *opens* are already in Now
(`Video opened`).

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Video closed | Desktop only; which video; which command when it is a command help video; whether playback reached the end | How often playback of an opened video reaches the end (with `Video opened`). |
| Selection options expanded | Desktop only; which command | Which written command’s Selection Options dropdowns get expanded. After a wrong selection, how often people expand the selection options versus try again with no help (with `Invalid selection`). |
| Invalid selection | Desktop only; which command; which variant if the command has variants | After a wrong selection, how often people expand selection options versus dismiss and try again with no help (no `Selection options expanded` before the next try). How often invalid selections are made. Which commands have a lot of invalid selections. |

### Drawing and commands

Covers the Later bullets under **Drawing and commands** (look/select-only
visits and invalid selections).

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Drawing saved | As in Owning drawings | Look/select-only visits: any pre-clear `mutationsSinceSave` above 0 means the visit changed the drawing (see working notes). |
| Visit ended | As in Visits | Look/select-only visits: `mutationsSinceSave` above 0 at leave means the visit changed the drawing; count at leave alone is not enough if they Saved earlier (use `Drawing saved` too). |
| Invalid selection | As in Help and videos | How often invalid selections are made. Which commands have a lot of invalid selections. |

### Command specific

Covers **Command specific** in the Later product questions.

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Command executed | As in Drawing and commands, plus when useful: which end or target (for example Grow), paste copy count and angle (for Paste) | Undo Grow then immediately Grow again on the other end. Multi-copy paste, then Undo, then change angle and paste again. |
| Command failed | Desktop only; which command; which variant if any; how it failed. Distinct from `Invalid selection` (bad pick before the command runs). **The set of failure reason labels is undecided for now** (Fillet first when we define them; others as we care) | In what ways Fillet fails. In what ways certain commands or variants fail when they do. |
| Setting changed | Desktop only; which setting (`infinite_rays`, or `ray_angle_mode` with values such as relative versus absolute/crosshairs); new value. Absolute/crosshairs is the ray-angle mode that locks to axes (including the common 90° case), not a separate momentary tool | How often infinite rays are turned on, and for how long (time until next off, or end of visit). Whether they are turned off right after a zoom command or right after another command. How often people switch relative versus absolute (crosshairs) ray angles. How often absolute/crosshairs mode is used. |
| Key used | Desktop only; which key; what it did (`close_dialog` or `unassigned` only). Do not emit this for keys that successfully fire a command — those are only `Command executed` with keystroke | How often Escape closes a dialog. How often keys we do not assign are used (for example DELETE). |

### Save and Drive

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Export PNG | Desktop only | How often people export a PNG. |

### General site info

| Event | Parameters | Questions it helps answer |
| --- | --- | --- |
| Invalid selection | As in Help and videos | Commands that have a lot of invalid selections. |
