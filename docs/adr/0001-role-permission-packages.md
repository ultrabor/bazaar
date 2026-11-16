# ADR 0001: Roles as permission packages copied to users

- **Status:** Accepted
- **Date:** 2026-11-14

## Context

Bazaar serves small companies. An owner creates employees and needs to control what each of them can do: create sales, see debts, edit products, assign roles.

Two needs pull in different directions:

1. Most employees with the same job need the same permissions. The owner should not set them one by one.
2. Some employees need an exception: one extra permission or one permission removed.

A pure role model (the user has a role, permissions are read from the role) handles the first need but not the second. A pure per-user model handles the second but makes the owner repeat the same setup for every employee.

## Decision

Permissions are stored per user in `user_permissions`. A role is a package of permissions in `role_permissions`.

- Permission names use the `object.action` format: `products.create`, `roles.assign`. Base actions are `create`, `view`, `edit` and `archive`. Other actions are added per object.
- The permission middleware checks only `user_permissions`. A request never reads the role.
- Assigning a role copies the role's package to the user. Changing the role first removes the permissions of the old package, then copies the new one. Both steps run in one transaction.
- Adding a permission to a role grants it to every user with that role. Removing it revokes it from those users. This also runs in one transaction.
- Individual permissions can be added or removed for one user with `POST /permission/set` and `POST /permission/remove`.
- A user can grant a permission, edit a role's package or assign a role only within their own permissions. Nobody can build a role wider than themselves.
- Each company has one system role, `owner`, with all permissions. It cannot be edited or assigned, and the owner cannot be moved to another role. This keeps at least one user with full access.

## Consequences

**Good**

- The permission check is one indexed lookup in `user_permissions`.
- Exceptions for a single employee are possible without creating a new role.
- Privilege escalation through roles is blocked by the subset check.

**Trade-offs**

- `user_permissions` does not record where a permission came from. When a permission is removed from a role, it is also removed from a user who had it as an individual exception. A `source` column (role or manual) fixes this and is planned after the MVP.
- Permission data is duplicated between roles and users. Every role change must update both tables inside one transaction.

## Alternatives considered

- **Read permissions from the role on each request.** Simpler, but no per-user exceptions.
- **Per-user permissions only, no roles.** Flexible, but the owner repeats the same setup for every employee.
