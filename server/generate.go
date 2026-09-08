// Package server anchors repository-owned generation commands.
package server

//go:generate go tool oapi-codegen -config codegen/oapi/public.yaml ../contracts/openapi/public.yaml
//go:generate go tool oapi-codegen -config codegen/oapi/admin.yaml ../contracts/openapi/admin.yaml
//go:generate go tool sqlc generate -f db/sqlc.yaml
