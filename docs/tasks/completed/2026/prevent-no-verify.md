# Prevent --no-verify

Is there any way for us to prevent the user from adding --no-verify to the commit command and pypassing the hook?

## Solution

1.  **Detection of --no-verify Bypass:** Updated `install-hook` to install both `pre-commit` and `post-commit` hooks. Since Git does not skip `post-commit` on `--no-verify`, we now run `taskhash check` in the post-commit phase and issue a visible warning if the developer bypassed the intended checks.
2.  **Robust Upgrades:** Fixed fragile string-based version parsing in `taskhash up` to use proper JSON parsing for the GitHub API. Added better error handling and cross-device update fallbacks.
3.  **Deployment Recommendations:** Recommended placing the binary in `tools/taskhash/taskhash` for shared repo consistency. Updated the `install` command and documented it in README.md.
4.  **Documentation:** Updated `README.md` and `GEMINI.md` to reflect new capabilities and recommended project structure.

---
**Completed in commit:** `<pending-commit-id>`
