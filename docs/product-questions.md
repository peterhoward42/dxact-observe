# Product questions

These are the product questions we want the observability system to answer. A
star (`*`) means we already send something like this to Vercel today. The events
and fields proposed to answer them are in [event-catalogue.md](./event-catalogue.md).

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

The matching Now/Later cut for what we emit is in
[event-catalogue.md](./event-catalogue.md).

## Now

### First visit

#### Arrival and examples

- On a first visit, how far does someone get through the newcomer steps (arriving, browsing examples, opening help, trying to create their own) before they stop doing anything we can see? *
- Do first visits arrive on a deep link to an example, or on the plain app address? *
- How long is a first visit, from the first thing we record to the last?
- When someone arrives on a deep link, do they stay on that drawing, or also open the list of other examples?
- On desktop, when someone arrives on the plain address, do they pick an example from the entry splash, or leave without picking?
- When newcomers open the list of example drawings, which ones do they pick? *
- How often do newcomers open that list and leave without picking any drawing? *

#### Desktop coaching on the first visit

- Once an example is on screen on desktop, do they take “How to move around in this drawing”? *
- Do they open the zoom video that note points at? *
- Do they open “How DrawExact works”? *
- Do they try “Create your own drawing”? *

We dropped “did they leave the desktop newcomer footer?” here: that footer only
clears when someone becomes an owner on this device, so it duplicates the
owning-drawings questions below.

### Coming back

#### Whether they return

- How often does someone who had a first-visit session open DrawExact again another day in the same browser?
- Of people who come back at least once, how long is the gap: next day, a few days, a week or more? Or same day?
- How many separate days do they come back over the first week, or the first month?
- How many people come back once and then disappear, versus keep coming back?
- What counts as a “regular” user for us? For example, someone who has opened the app on three different days, and how many first-time visitors reach that?
- After someone has saved or opened their own drawing, how often do they keep coming back?

#### What a return visit looks like

- When someone returns, do they arrive on a deep link again, or on the plain address?
- On mobile, how many returning visitors still see the newcomer start screen?
- On desktop, how many returning visitors still see the newcomer footer (this device is not yet an owner)?
- On a return visit, do they mostly look at examples again, open help, or actually draw and save?

### Becoming someone who owns drawings

- How often do people see the Connect Google Drive ask and leave without connecting?
- How many first-visitors connect Drive?
  - How many people return on a later visit and open an owned drawing?
- How many users who connect Google Drive go on to create their first owned drawing, either by starting a blank drawing or saving a copy of an example? How many choose each route?
- How often does someone’s first owned drawing come from saving an example they were looking at?

### Help + Videos

#### Videos

- In what order do different commands’ help videos get opened?
- For a given command, is the longer guide or the short reminder opened first? *
- How many times is each command help video opened?
- Do people open a command’s videos before using that command?

#### Written help

- Do people open any command’s written help?
- For a given command, do people open its help before they’ve used the command, or only after?
- After someone opens a command’s written help, how often do they then run that command in the same visit?

### Drawing and commands

- How often do each of the commands and their variants get used?
  - Are there some variants that are rarely used?
- Do people use Undo after some commands more than others?
- How often do commands get fired by keystroke vs by clicking on the command strip?

### Save and Drive

- How often do people open an existing drawing of their own, versus opening an example — counting both picks from the list or splash, and arrivals on a deep link to an example?
- How many unsaved changes do people build up before saving?
- How often do users leave DrawExact with unsaved changes, and how many changes do they leave behind?

### Mobile Users

#### First Viewer visit

- How many first Viewer visits arrive on a deep link to an example, or on the plain app address?
- On a first Viewer visit, how far do people get (start screen, open the examples list, pick a drawing, open “How DrawExact works”, try “Create your own drawing” (the handoff that points them at a computer) ) before we stop getting any further steps from that visit?
- How often do people tap “Show me some drawings” versus “How DrawExact works” on the newcomer start screen?
- How often do people tap “Create your own drawing” on the Viewer (the path that points them at making drawings on a computer)?

#### Share and handoff

- How often do people share a drawing?
- On a phone or tablet Viewer visit, do they copy the link that hands the drawing to a computer? *
- When they share, was it from a deep-linked example, or from something they picked from the list?

#### Own drawings on the phone

- How often do people tap “I already have drawings” or “Your drawings”, and then actually connect Google Drive?
- Once Drive is connected on a phone, how often do they open one of their own drawings, versus bounce after Connect?
- How often does someone open their own drawing on the Viewer at all?

#### Coming back on mobile

- Of people who only ever use mobile in that browser, how many later come back another day?
- Do mobile visits mostly stay as looking at examples, or do they lead to share / Connect / open-own more often over time?

### General Site Info

- How many different browsers visit in a week?
- Of everything we record, how much is first-visit behaviour, how much is help, and how much is real drawing?
- Are there commands that rarely get used?
- Help windows or videos that rarely get opened? (orientation, move-around, command written help, newcomer zoom, command help videos)
- Commands whose written help get opened a lot?
- Commands whose help videos get opened a lot?
- How long does a typical DrawExact session last, from the first thing we record in that visit to the last?

## Later

### Help + Videos

#### Videos

- How often does playback of an opened video reach the end?

#### Written help

- Which written command’s Selection Options dropdowns get expanded?
- After a wrong selection, how often do people expand the selection options, versus dismiss and try again with no help?

### Drawing and commands

- How often does a drawing session include only looking and selecting, with no command that changes the drawing — meaning the Save toolbar unsaved-changes count (`mutationsSinceSave`) never left 0 in that visit?
- How often are invalid selections made?

### Command specific

- Perhaps track in what ways Fillet fails?
- How often do users turn on infinite rays? For how long?
  - Do they toggle them off immediately after changing zoom?
  - Do they toggle them off immediately after using another command?
- How often do users switch between relative and absolute (crosshairs) ray-angle modes? How often is absolute/crosshairs mode used?
- In what ways do certain commands/variants fail (when they do)
- How often does a user undo a Grow, then immediately reuse Grow by clicking on the other end?
- How often do people make a multi-copies paste, then undo, then change the angle, and repaste?
- How often does Escape get used to close a dialog window?
- How often do keys we don’t assign get used, e.g. DELETE?

### Save and Drive

- How often do people export a PNG?

### General Site Info

- Commands that have a lot of invalid selections made?
