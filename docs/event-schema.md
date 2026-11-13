# Event schema

- The events currently emitted from dxact-draw provide a conceptual incumbent 
  event schema that points us in roughly the right direction.
- But this should probably changed according to the following pressures:
	- They have been unduly influenced by Vercel's needs and API
	- Pete finds their names and parameters frankly difficult to understand
	  (despite guiding their formation). They should be much easier to grasp.
	- I think we might need some kind of schema versioning to cope with changes
	  down the line - not sure - needs some analysis

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
	  storage.

