---
name: auth-security
description: 'Project rules for authentication and session handling (AD-005). Use when planning, building or reviewing login, logout, sessions, password storage, user deactivation or role changes in the users/rbac features.'
---

# Auth security

AD-005 is the constraint: local argon2id passwords, opaque server-side session token in an
httpOnly, SameSite=Lax cookie, revocable. No JWT, no signed cookies, no OIDC. These rules close
the gaps AD-005 leaves open; record each as a one-way door or check in the `users` plan.

1. Session token: 32 bytes from `crypto/rand`, sent raw in the cookie, stored only as its SHA-256
   hash. A database leak must not yield usable tokens.
2. Issue a new token on every successful login; never reuse a token that existed before
   authentication.
3. Unknown user and wrong password take the same path: run an argon2id verify against a fixed
   dummy hash, return the same Problem Details body and status. No user enumeration by message or
   timing.
4. Pin one argon2id parameter set (memory, iterations, parallelism, salt and key length) in code
   and encode it in the stored hash so it can be raised later.
5. Cookie `Secure` flag comes from config (off only for local dev over http).
6. Deactivating a user or changing their roles deletes their sessions in the same transaction, so
   the next request is rejected (AD-005, AD-007).
7. Rate-limit login attempts per account and per client IP.
8. Audit rows take the client IP from a forwarding header only when the request came through a
   configured trusted proxy; otherwise use the connection address.
