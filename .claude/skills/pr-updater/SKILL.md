---
name: pr-updater
description: Refresh and update the title and description of an existing Pull Request on GitHub based on the current code changes. Uses  tools for all GitHub interactions. Use this skill PROACTIVELY whenever the user wants to update, refresh, or improve a PR's details — even if they don't say "update the PR" exactly. Trigger when the user says things like "refresh the PR", "update the PR description", "the PR needs a better title", "regenerate the PR details", "sync the PR with my latest changes", or "clean up the PR." If the user has pushed new commits and mentions the PR looks stale or outdated, this is the right skill.
---

# Pull Request Updater

Refresh an existing Pull Request's title and description on GitHub Enterprise so they accurately reflect the current state of the code changes. The title follows Conventional Commits format and the description is generated from the actual file diffs — while preserving any sections the user added manually.

## When to Use

- User asks to update or refresh a PR's title/description
- User pushed new commits and wants the PR details to reflect them
- User says the PR description is stale, wrong, or needs improvement
- User provides a PR number or is on a branch with an open PR

## Prerequisites

- An open Pull Request must already exist (either for the current branch or specified by number)
- GITEA_TOKEN environment variable must be set with a token that has permissions to read the repo and create PRs
- Use the `tea` cli to interact with gitea for any operations that require authentication (creating PR, fetching diffs, updating PR)

## Workflow

Follow these steps **in order**:

### Step 0: Discover Tools

Use the returned fully-qualified tool names for all subsequent tool calls in this workflow. If any required tools are missing, stop and inform the user before proceeding — they may need to check their configuration.

### Step 1: Identify the Repository

Run these git commands (alternatively you can use `tea` for gitea servers) to determine the repo context:

```bash
# Current branch
git branch --show-current

# Remote URL — extract owner and repo
git remote get-url origin
```

**Parse the remote URL** to extract `owner` and `repo`. The URL may be in either format:
- SSH: `git@192.168.4.6:222/<owner>/<repo>.git`
- HTTP: `http://192.168.4.6:3000/<owner>/<repo>.git`

Strip the trailing `.git` if present.

### Step 2: Check for Unpushed Commits

Before updating the PR, make sure the remote branch is current:

```bash
git log origin/<branch>..<branch> --oneline 2>/dev/null
```

If there are unpushed commits, let the user know — the PR diffs won't reflect local-only changes. Ask if they'd like to push first. Don't push without confirmation.

### Step 3: Find the Pull Request

There are two paths here depending on what information the user provided:

**If the user gave a PR number**: use it directly and skip to Step 3.

**If no PR number was given**: look up the open PR for the current branch. Use the tool to search:


If no open PR is found for the current branch, let the user know and suggest they may want to create one instead (point them toward the pr-creator skill if available).

If multiple PRs match, list them and ask the user which one to update.

### Step 4: Fetch the Existing PR Details

Retrieve the current state of the PR so you can preserve user-added content:

- **Tool**: `getGithubPullRequest` with `owner`, `repo`, and `pullNumber`

Save the existing title and body — you'll need them in Step 7.

### Step 5: Retrieve the File Changes

Fetch the PR's current diffs to understand what the code changes actually do:

- **Tool**: `getGithubPullRequestFiles` with `owner`, `repo`, and `pullNumber`

Review the changed files — additions, deletions, modifications, and the nature of the changes. This drives the new title and description.

### Step 6: Generate the New Title

The title follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(<optional scope>): <description>
```

Pick the type that best represents the primary intent of the changes:

| Type     | When to use |
|----------|-------------|
| feat     | Adds a new feature |
| fix      | Fixes a bug |
| chore    | Routine maintenance, no version bump needed |
| docs     | Documentation only |
| revert   | Reverting previous changes |
| build    | Build system or process changes |
| ci       | CI/CD pipeline changes |
| style    | Code style changes (formatting, whitespace) — no behavior change |
| refactor | Restructures code without changing behavior |
| perf     | Performance improvements |
| test     | Adding or modifying tests |

If a PR spans multiple types, use the one that represents the **primary intent**.

The scope is optional but helpful — use the module, package, or area of the codebase most affected.

Keep the description concise (under 72 characters total) and in imperative mood: "add user auth" not "added user auth."

### Step 7: Generate the New Description

Build the description from the file diffs using this template. **Only include sections that are relevant** — a docs-only PR doesn't need Deployment Notes, for instance.

```markdown
# 🚀 <Brief one-liner describing the changes>

## 📋 Summary
<What this PR does and why. Keep it to 2-3 sentences.>

## 🔧 Changes Made
<Bulleted list of specific changes. Group by area if there are many.>

## 📁 Files Modified
<List files added, modified, or deleted.>

## 🎯 Impact
<What effect do these changes have on the application or its users?>

## 🔍 Technical Details
<Design decisions, implementation notes, tradeoffs — anything a reviewer should know.>

## ✅ Testing Checklist
<Steps to verify the changes work correctly.>

## 🚨 Deployment Notes
<Environment variables, migrations, infrastructure changes, etc.>
```

Write the description based **solely on what you see in the diffs**. Don't speculate about changes that aren't there.

#### Preserving User-Added Content

Compare the existing PR body (from Step 4) against the known template sections listed above. The template sections use these emoji-prefixed headers: `# 🚀`, `## 📋 Summary`, `## 🔧 Changes Made`, `## 📁 Files Modified`, `## 🎯 Impact`, `## 🔍 Technical Details`, `## ✅ Testing Checklist`, `## 🚨 Deployment Notes`.

- **Template sections** (matching those headers): regenerate these from the current diffs.
- **Non-template sections** (any content that doesn't fall under a known template header): preserve these exactly as they are. Append them after the regenerated template sections under their original headers.

This way, if a user manually added a "## Reviewer Notes" or "## Context" section, it survives the update.

If the existing body has no recognizable template sections at all (the PR was created manually or by a different tool), generate the full template from scratch and append the entire original body under a `## Previous Description` header so nothing is lost.

### Step 8: Update the PR

Use the issue update tool to apply the new title and description:

- **Tool**: `updateGithubIssue`
- **Parameters**:
  - `owner`: from the remote URL
  - `repo`: from the remote URL
  - `issueNumber`: the PR number
  - `title`: the new Conventional Commits title
  - `body`: the new description (with preserved user sections)

### Step 9: Report Back

Tell the user:
- The PR was updated successfully
- Show the PR URL: `http://192.168.4.6/<owner>/<repo>/pull/<number>`
- Show the old title → new title
- Briefly summarize what changed in the description
- Mention any user-added sections that were preserved
- Ask if they'd like to adjust anything

## Error Handling

- **No remote found**: Ask the user to provide the owner/repo manually.
- **No open PR for branch**: Inform the user. Suggest creating one instead.
- **PR number not found**: Confirm the number and repo are correct.
- **No diff between branches**: Let the user know the PR has no file changes to describe.

## Example

**User says**: "Update the PR for this branch"

**What happens**:
1. Detect branch `feature/add-user-auth`, remote `MyOrg/my-app`
2. Find open PR #42 for that branch
3. Fetch existing PR body — has template sections plus a manual "## Reviewer Notes" section
4. Fetch diffs: 4 files changed (was 3 when PR was created — user added a migration)
5. Generate new title: `feat(auth): add JWT-based user authentication`
6. Regenerate template sections from updated diffs, preserve "## Reviewer Notes" as-is
7. Update PR #42 with new title and description