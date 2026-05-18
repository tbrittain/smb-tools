# Architecture

Design decisions and guiding principles for the smb-tools rewrite. These documents capture the *why* behind structural choices so that future work stays coherent with the original intent.

| File | Contents |
|------|----------|
| [decisions.md](decisions.md) | Committed technology choices with rationale |
| [backend-structure.md](backend-structure.md) | Go package layout and idiomatic patterns |
| [data-layer.md](data-layer.md) | SQLite strategy, per-franchise DB design, two connections, migrations, schema principles |
| [testing-strategy.md](testing-strategy.md) | Testability requirements and approach at each layer |
| [ux-flows.md](ux-flows.md) | Core user-facing flows: sync, franchise switching, legacy migration, CSV export |
| [snapshot-strategy.md](snapshot-strategy.md) | Save game snapshot persistence, deduplication, compression, storage management |
| [open-decisions.md](open-decisions.md) | Pending decisions that require further discussion before implementation |

## Reading Order

Start with `decisions.md` for the committed stack, then `backend-structure.md` and `data-layer.md` for how the Go side is organized, then `testing-strategy.md` for how testability is enforced throughout. Check `open-decisions.md` before starting any frontend work.
