---
name: pr-creator
description: Create a Pull Request on GitHub from the current branch, then automatically generate a polished title and description based on the actual code changes. Uses gh Server tools for all GitHub interactions. Use this skill PROACTIVELY whenever the user wants to create a PR, open a pull request, submit changes for review, or push their work up for merging — even if they don't explicitly say "create a PR." Also use it when the user says things like "I'm done with this feature," "let's get this reviewed," "open a PR for this," "submit this," or "send this up." If the user is working on a branch that isn't the default branch and indicates they're finished with their changes, this skill is likely what they need.
---

# Pull Request Creator

Create a Pull Request on GitHub Enterprise and automatically populate it with a well-structured title (Conventional Commits) and description derived from the actual file changes.

## When to Use

- User asks to create/open a PR
- User says they're done with changes and want them reviewed
- User wants to submit work from a feature branch
- User invokes `/create-pr` or similar

## Prerequisites

- The current branch must have commits ahead of the base branch
- GITEA_TOKEN environment variable must be set with a token that has permissions to read the repo and create PRs
- Use the `tea` cli to interact with gitea for any operations that require authentication (creating PR, fetching diffs, updating PR)

## Workflow

Follow these steps **in order**:

### Step 0: Discover Tools

Use the returned fully-qualified tool names for all subsequent tool calls in this workflow. If any required tools are missing, stop and inform the user before proceeding — they may need to check their server configuration.

### Step 1: Gather Context from the Local Repo

Run these git commands (alternatively you can use `tea` for gitea servers) to collect the information you need:

```bash
# Current branch (this becomes the PR's head branch)
git branch --show-current

# Remote URL — extract owner and repo from it
git remote get-url origin

# Verify there are commits to push
git status
```

**Parse the remote URL** to extract `owner` and `repo`. The URL may be in either format:
- SSH: `git@192.168.4.6:222/<owner>/<repo>.git`
- HTTP: `https://192.168.4.6:3000/<owner>/<repo>.git`

Strip the trailing `.git` if present.

### Step 2: Ensure Commits Are Pushed

Check whether the local branch has been pushed to the remote:

```bash
git log origin/<branch>..<branch> --oneline 2>/dev/null
```

If there are unpushed commits, **ask the user** if they'd like you to push first. Do not push without confirmation. If the remote branch doesn't exist at all, the user definitely needs to push — let them know.

### Step 3: Detect the Base Branch

Use the gh tool to get the repository's default branch:

- **Tool**: `getGithubRepository` with the `owner` and `repo` values
- The response contains a `default_branch` field — use that as the base branch

This matters because not every repo uses `main`. Some use `master`, `develop`, or something custom. Always detect rather than assume.

### Step 4: Create the Pull Request

Use the gh tool to create the PR:

- **Tool**: `createGithubPullRequest`
- **Parameters**:
  - `owner`: from the remote URL
  - `repo`: from the remote URL
  - `title`: use a short placeholder like `"WIP: <branch-name>"` — you'll replace this in a moment
  - `body`: `"Generating description from file changes..."`
  - `head`: current branch name
  - `base`: the default branch from Step 3
  - `draft`: `false` (create as ready for review)

If creation fails because there are no differences between head and base, tell the user there's nothing to open a PR for.

### Step 5: Retrieve the File Changes

Now that the PR exists, fetch its diffs:

- **Tool**: `getGithubPullRequestFiles` with `owner`, `repo`, and `pullNumber` from the creation response

Review the changed files — look at the diffs, file additions/deletions, and the nature of the changes. This is what drives the title and description.

### Step 6: Generate the PR Title

The title MUST follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification:

```
<type>(<optional scope>): <description>
```

**There is no colon between the type and the scope parentheses.**

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

The scope is optional but helpful — use the module, package, or area of the codebase that's most affected.

The description should be concise (ideally under 72 characters total for the whole title) and written in imperative mood: "add user auth" not "added user auth."

### Step 7: Generate the PR Description

Build the description from the file changes you reviewed. Use this template, but **only include sections that are relevant** — omit anything that doesn't apply. A docs-only PR doesn't need Deployment Notes, for instance.

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

### Step 8: Update the PR

GitHub's API treats Pull Requests as Issues for certain write operations. Use the issue update tool to set the final title and description:

- **Tool**: `updateGithubIssue`
- **Parameters**:
  - `owner`: from the remote URL
  - `repo`: from the remote URL
  - `issueNumber`: the PR number from Step 4
  - `title`: your generated Conventional Commits title
  - `body`: your generated description

### Step 9: Report Back

Tell the user:
- The PR was created successfully
- Show the PR URL (format: `https://192.168.4.6/<owner>/<repo>/pull/<number>`)
- Show the title you generated
- Briefly summarize the description
- Ask if they'd like to adjust anything (title, description, add reviewers, convert to draft, etc.)

## Error Handling

- **No remote found**: Tell the user you can't detect the repository. Ask them to provide the owner/repo.
- **Branch not pushed**: Suggest they push first, offer to do it if they confirm.
- **PR already exists for this branch**: Inform the user. Offer to update the existing PR's title and description instead.
- **No diff between branches**: Let the user know there are no changes to create a PR for.

## Example

**User says**: "Create a PR for my changes"

**What happens**:
1. Detect branch `feature/add-user-auth`, remote `MyOrg/my-app`
2. Get default branch → `main`
3. Create PR #42 with placeholder title
4. Fetch diffs: 3 files changed — new auth middleware, updated routes, new tests
5. Generate title: `feat(auth): add JWT-based user authentication`
6. Generate description with Summary, Changes Made, Files Modified, Impact, and Testing Checklist sections
7. Update PR #42 with final title and description