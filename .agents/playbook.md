# Engineering playbook

1. Confirm branch/status/history and preserve existing work. Read the active spec
   first, then only required/affected context and contracts. Audit the implementation
   and establish the available verification baseline.
2. Implement the smallest complete change. Keep unused dependencies and future
   domain scaffolding out. Change source specifications before generating code.
3. Exercise observable behavior, timeouts, failure handling and relevant integration
   boundaries. Never log driver errors containing connection details or credentials.
4. Use `pnpm check` for final repository verification. Run explicit migration, real
   driver smoke and image checks required by the active spec. CI uses the same check
   logic with disposable services; `.local` files are never CI inputs.
5. Regenerate again, check drift (including untracked generated files), review the
   complete final diff and dependency changes, and audit tracked content for secrets.
6. Update `CHANGELOG.md`. Record ADRs only for significant long-lived decisions.
7. Stage the coherent change, review the staged diff, commit locally on `dev` only
   after acceptance, and verify status. Report commands actually executed, results,
   unresolved blockers and SHA. No push, merge, tag or release without instruction.

Stop and report evidence for consequential conflicts: unsupported Fiber v3 codegen,
River custom-schema/privilege failure, tracked secrets, required shared Infra admin
access, prohibited components or incompatible selected toolchains. Do not work around
these by changing architecture. Routine implementation details need no new approval.
