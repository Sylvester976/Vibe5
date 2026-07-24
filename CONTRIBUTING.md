# Contributing

Thanks for picking this up. A few norms to keep the history and codebase clean.

## Commits

- Write commits as yourself. Do not include `Co-authored-by: Claude` or any
  AI-attribution trailer in commit messages, regardless of what tooling you used
  to help write the code — commit authorship reflects the person responsible for
  the change, not the tools that assisted.
- Keep commits scoped to one logical change. Reference the relevant
  `docs/ARCHITECTURE.md` section in the body if the change touches a design
  decision documented there.
- Conventional commit prefixes are welcome but not required: `feat:`, `fix:`,
  `refactor:`, `docs:`.

## Before opening a PR

- Read `docs/ARCHITECTURE.md` — especially §8 (open decisions) — a PR that
  contradicts a decision made there should say why in the description.
- If you're adding a new Spotify API call, confirm the endpoint is still
  available to Development Mode apps (see README's note on API limits) before
  building around it.
- UI changes should stay inside the design tokens in `docs/ARCHITECTURE.md` §6
  (color, type, layout, signature element) rather than introducing new ad hoc
  values — open an issue first if you think a token needs to change.

## Code style

- Go: `gofmt` + `go vet` clean, no linter warnings ignored without a comment.
- React: functional components, hooks, no class components.
