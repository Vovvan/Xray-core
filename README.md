# mixed-usermanager patches

Patch archive for the dynamic `UserManager` support on `Mixed`/SOCKS in
`XTLS/Xray-core`. The base release is recorded in `PATCH-BASE`.

The series intentionally contains only the proxy and test changes. Fork-only CI
changes are excluded so the patches stay focused and remain applicable to
pristine upstream.

## Apply to pristine upstream

```sh
git clone --branch v26.10.21 https://github.com/XTLS/Xray-core.git xray-core
./scripts/apply.sh xray-core
```

Or apply manually:

```sh
git -C xray-core am -3 "$PWD"/patches/*.patch
```

The full test suite needs `resources/geoip.dat` and `resources/geosite.dat`.
They can be downloaded from `Loyalsoldier/v2ray-rules-dat`.

## Regenerate from the feature branch

```sh
./scripts/export.sh /path/to/xray-core-feature v26.10.21
```

The export command excludes `.github` changes and rewrites `series`. Update
`PATCH-BASE` if the upstream base changes.

## Rebase flow

1. Rebase the feature branch onto the new upstream tag.
2. Run `scripts/export.sh` against the rebased branch.
3. Update `PATCH-BASE`.
4. Dispatch `patch-verify` with the new base and `patches/mixed-usermanager`.
