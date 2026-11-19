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
  Event names and parameters are still undecided. Phase 1 is to derive those
  from the questions. The emitting contract is owned by dxact-draw (see
  coupling-to-other-dxact-repos.md).

## Common event envelope

Every event carries:

- Event type — what happened
- Timestamp — when it happened, in UTC
- User identifier — browser-installation UUID (see below)
- Schema version — so historical records can be recognised (see below)
- Additional parameters — information particular to that event, as arbitrary
  JSON, so individual events can gain fields without continually restructuring
  the whole schema

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

# Identifying users

- We didn't need to identify users with the Vercel telemetry because it did that
  for us.
- We will now do that ourselves. 
- Each event will carry a proxy user id minted by the sender
	- A unique string that anonamises the user but serves as a unique key for
	  that user.
	- Use a UUID or something similar that is shorter but sufficient
	- We don't need this ID to also encapsulate a timestamp but don't care 
	  if it does.
	- We won't attempt to identify real unique users. As our proxy we will mint a
	  new ID when we first encounter drawexact running on any given
	  device/browser combination.
	- We persist that value for subsequent use in that Browser's persistent 
	  storage (localStorage).
- The UUID identifies a browser installation, not a person across devices. That
  is an accepted compromise; recognising the same person across devices would
  need more infrastructure and potentially account identity.
