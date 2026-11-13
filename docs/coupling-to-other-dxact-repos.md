# Coupling to other DrawExact repositories

- DrawExact is spread across several repositories, all with names that begin with
  dxact_
- These are intended to be organised into the developers filesystem under a
  common root directory
- The common parent is a convenience mainly for tooling and the odd documentation
  convenience, and should be used only as much as necessary.

- The telemetry events that this service will consume are emitted by code in the
  dxact-draw repository, and it is that repo that owns the event schema.
- The code in this repo is obliged to keep track of that schema as a slave.
- Probably reducing the mechanicl contract to a simple, and apprporiate  
  string based parity
