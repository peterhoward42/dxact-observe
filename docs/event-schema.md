# Event schema

- The events currently emitted from dxact-draw provide a conceptual incumbent 
  event schema that points us in roughly the right direction.
- But this should probably changed according to the following pressures:
	- They have been unduly influenced by Vercel's needs and API
	- Vercel Analytics is oriented toward conventional websites and page
	  visits. DrawExact is effectively a single-page application; most
	  meaningful interactions happen inside the app rather than through
	  navigation between pages, so that model is not necessarily a good fit.
	- Pete finds their names and parameters frankly difficult to understand
	  (despite guiding their formation). They should be much easier to grasp.
	- I think we might need some kind of schema versioning to cope with changes
	  down the line - not sure - needs some analysis

## Design around questions

- Design telemetry around the questions we might eventually want to answer,
  then identify the data required to answer them.
- Prefer a small set of clear, self-explanatory generic event types with
  qualifying parameters, rather than a separate event type for every concrete
  action. For example: one "command executed" event with the command name as a
  parameter, rather than one event type per command.
- Illustrative types only (not an agreed catalogue): application opened;
  command executed (which command); help opened (which resource); video viewed
  (which video; short or long).
- The product questions are in [product-questions.md](./product-questions.md).
  The draft event names and parameters are in
  [event-catalogue.md](./event-catalogue.md). They are not final until we say
  we are willing to implement that list. The emitting contract is owned by
  dxact-draw (see coupling-to-other-dxact-repos.md).

## Common event envelope

One JSON object per POST body (`Content-Type: application/json`). After
validation, storage writes a compact single-line re-encoding of the **accepted
envelope** (including `params`), not the raw POST body. Key order does not
matter on ingest; when writing the stored object, put `time` first so a
single-line object is easy to skim.

```json
{"time":"2026-11-24T08:45:12.345Z","type":"Application opened","browser_uuid":"8e6a4c1d-5f29-4e83-9e10-42bd735af0c2","schema_version":1,"params":{"face":"desktop","arrival":"plain"}}
```

| Field | Meaning and rules |
| --- | --- |
| `time` | When it happened, UTC. RFC 3339 with a `Z` suffix; fractional seconds allowed. Reject missing `Z` or non-UTC offsets. This payload timestamp is authoritative for reconstructing activity. It is not the GCS object's upload or creation time; network delays and retries can make those differ. |
| `type` | What happened. Non-empty string. Catalogue names as written (for example `Application opened`). |
| `browser_uuid` | Persistent pseudonymous browser-installation UUID (see Identifying users). Standard `8-4-4-4-12` hex form. Minted by the sender; ingest only checks form. |
| `schema_version` | JSON number so historical records can be recognised (see Schema versioning). Only `1` is accepted for now. |
| `params` | Information particular to that event, as a JSON object (may be `{}`). Individual events can gain fields here without restructuring the envelope. Envelope validation does not inspect keys inside `params`. |

Reject top-level arrays, multiple events in one body, missing envelope keys,
unknown top-level keys, or a `params` value that is not an object.

The fresh UUID used in the GCS object name (`events/YYYY-MM-DD/<event-uuid>.json`)
is a storage/naming concern on the ingest side. It is not in this envelope, and
the emit contract does not supply it.

Fractional seconds are allowed so actions that fall in the same clock second
can still sort into a plausible order when reconstructing a visit — for
example opening the examples list and immediately choosing an example. That
is not a hard guarantee: two events can still share the same timestamp. If
unambiguous ordering of events with identical timestamps becomes necessary,
a per-user sequence number can be added later.

## Schema versioning

- When the format changes, increment the schema version on new events.
- Do not rewrite historical telemetry objects simply because the schema moved.
- A version-aware loader recognises earlier versions and converts them into the
  current in-memory data model when loading.
- Obsolete fields can be ignored during that construction; historical records
  stay as stored.

## Schema coupling

- We will need at some stage to create a new event schema definition in the
  emitting repo. The design and documentation of the schema to be adopted in this
  repo may point to that as what it is intending to match, but probably not have
  a more direct coupling than that

# Identifying browsers (browser UUID)

- We didn't need a browser UUID with the Vercel telemetry because it did that
  for us.
- We will now do that ourselves. The envelope field is `browser_uuid`.
- Each event carries a persistent pseudonymous browser UUID minted by the
  sender. This is not a verified personal identity.
	- A unique string that keys one browser/device environment.
	- Use a UUID (standard `8-4-4-4-12` hex form on the wire).
	- We don't need this ID to also encapsulate a timestamp but don't care
	  if it does.
	- We won't attempt to identify real unique people. DrawExact mints a new
	  browser UUID on first use in a given browser/device environment and
	  reuses it on subsequent visits, subject to the persistence of local
	  browser storage (localStorage).
- The browser UUID identifies a browser installation, not a person across
  devices. That is an accepted compromise; recognising the same person across
  devices would need more infrastructure and potentially account identity.
