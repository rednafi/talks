# Context propagation in Go

A 40-minute talk built with [Slidev](https://sli.dev) and the
[tahta](https://github.com/zcag/tahta) `atelier` variant.

```sh
pnpm install
pnpm dev       # http://localhost:3030, presenter view at /presenter
pnpm build     # static site in dist/
pnpm export    # PDF (needs playwright-chromium)
pnpm lint      # tahta-lint
```

- `slides.md`: the deck and the speaker notes
- `components/`: `Code` and `Mark` (arrows to code lines), `TwoLevels`
  (diagram), and overrides of tahta's `Foot` (section slides aren't counted) and
  `Fit`
- `examples/`: runnable Go that checks every claim the slides make
  (`go run ./examples`) and compiles every slide snippet (`go vet ./...`)
- `scripts/shots.mjs`: screenshots of each slide from the dev server

Gopher art by Egon Elbre (CC0), after the Go gopher by Renee French.
