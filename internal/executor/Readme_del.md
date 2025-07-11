## Adding a New Health Check

To add a new health check in Dr-HPC, follow the structured approach below. The goal is to maintain consistency across all components: Python, Bash, and Go-based implementations.

---

### 🧩 Step 1: Define the Check Name

Choose a unique and descriptive name for the new check. Use this name consistently across all implementations:
- `gid_index_check.py`
- `gid_index_check.sh`
- `gid_index_check.go`

---

### 🐍 Python & Bash Standalone Scripts

These scripts are typically **shape-specific** and are used for standalone validations or ad-hoc runs.

1. In the Dr-HPC repo, locate an existing check (e.g., `run_gid_index_check`).
2. Refactor the check into a self-contained Python script (`gid_index_check.py`).
3. Convert the logic into an equivalent Bash script (`gid_index_check.sh`) if needed.

> These scripts do **not need to be shape-agnostic** and can include shape-specific logic inline.

---

### 🔧 Go Framework Integration

The Go framework is designed to be **shape-agnostic** and integrated into the main Dr-HPC runtime.

1. **Reuse the existing Go framework**:
    - The framework supports multi-shape execution and centralized control.

2. **Update the executor package**:
    - Determine the actual command run by `run_gid_index_check`.
    - Add the command to the `executor` package for unified command execution.

3. **Add test parameters to `test_limits`**:
    - Identify the required thresholds or limits for the check.
    - Add these to the `test_limits` JSON configuration.

4. **Create the test implementation file** (`gid_index_check.go`):
    - Read threshold values from the `test_limits` package.
    - Use the `executor` package to run the check.
    - Parse and validate the output.
    - Compare results against the threshold values.
    - Report success/failure to the `reporter` package.

5. **Update the `reporter` package**:
    - Extend it to handle and display results from the new `gid_index_check.go`.

6. **Update `recommendations.json`**:
    - Add user-facing guidance for scenarios where `gid_index_check.go` reports a failure.

7. **Update the `recommender` package**:
    - Ensure it reads the new recommendation and displays relevant information in the console output.

8. **Register the new test in the `cmd` package**:
    - Since `gid_index_check` is a **Level 1 test**, add it to the `level1.go` file within the `cmd` package.
    - This ensures the test is executed as part of the standard Level 1 test suite and is exposed through the CLI entry point.

---

### ✅ Final Checklist

- [ ] Consistent naming across `.py`, `.sh`, and `.go` files, `test_limits.json` and `recommendations.json` fields
- [ ] Refactored Python/Bash scripts are self-contained
- [ ] New Go check integrated into `executor`, `test_limits`, `reporter`, and `recommender`
- [ ] Validations and thresholds are centrally defined
- [ ] Recommendations for user action are documented and surfaced
- [ ] `gid_index_check` registered in `cmd/level1.go` under Level 1 tests
---
