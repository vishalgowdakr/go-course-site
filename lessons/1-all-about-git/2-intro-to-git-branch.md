# Git Branch

## Introduction to Git Branches

Branches are a powerful feature in Git that allow you to isolate development work. They are lightweight and easy to create, making them ideal for experimenting with new features or fixing bugs without affecting the main codebase.

### Creating a Branch

To create a new branch, use the `git branch` command followed by the branch name:

```bash
git branch <branch_name>
```

This will create a new branch with the specified name. To switch to the new branch, use the `git checkout` command:

```bash
git checkout <branch_name>
```

### Merging Branches

Once you have made changes on a branch and are ready to integrate them into another branch, you can use the `git merge` command. First, switch to the target branch where you want to merge the changes:

```bash
git checkout <target_branch>
```

Then, merge the changes from the source branch into the target branch:

```bash
git merge <source_branch>
```
