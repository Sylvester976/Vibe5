// Package migrations embeds the SQL migration files so they ship inside the
// compiled binary instead of depending on files being present at runtime.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
