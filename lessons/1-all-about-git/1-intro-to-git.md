# Git Quick Start Guide for Experienced Developers

**Introduction**

Git is a powerful version control system (VCS) essential for modern software development. It allows developers to track changes in code, collaborate effectively, revert to previous versions, and manage codebases efficiently. 

**Installation and Setup**

**Prerequisites:**

* A terminal application (e.g., Bash, Zsh)

**Installation:**

1. Download and install Git for your operating system from the official website: [https://git-scm.com/downloads](https://git-scm.com/downloads).
2. Verify installation by running `git --version` in your terminal. 

**Basic Commands and Code Snippets**

1. **Initializing a Repository:**
   - Creates a new Git repository in the current directory.
   ```bash
   git init
   ```

2. **Adding Files:**
   - Adds specific files to the staging area for version control.
   ```bash
   git add <filename>
   ```
   - Adds all modified and untracked files in the working directory.
   ```bash
   git add .
   ```

3. **Committing Changes:**
   - Captures the current state of the staged files with a descriptive message.
   ```bash
   git commit -m "<commit message>"
   ```

4. **Viewing Commit History:**
   - Lists all commits made to the repository.
   ```bash
   git log
   ```

5. **Viewing File Changes:**
   - Shows the difference between the working directory and the last commit (or a specific commit).
   ```bash
   git diff <filename>  # For a specific file
   git diff             # For all staged changes
   git diff HEAD~1      # Compare with the previous commit
   ```

6. **Branching:**
   - Creates a new branch to isolate development work.
   ```bash
   git branch <branch_name>
   ```
   - Switches to an existing branch.
   ```bash
   git checkout <branch_name>
   ```

7. **Merging Branches:**
   - Integrates changes from one branch into another.
   ```bash
   git checkout <target_branch>
   git merge <source_branch>
   ```

8. **Remote Repositories (Using GitHub):**
   - Creates a remote repository on GitHub. (**Outside Git commands**)
   - Clones an existing remote repository.
   ```bash
   git clone <remote_repository_url>
   ```
   - Pushes local commits to the remote repository.
   ```bash
   git push origin <branch_name>
   ```
   - Pulls changes from the remote repository to your local branch.
   ```bash
   git pull origin <branch_name>
   ```

**Advanced Usage**

* **Undoing Changes:**
   - Revert the last commit (careful, loses changes).
   ```bash
   git revert HEAD
   ```
   - Discard untracked changes in the working directory.
   ```bash
   git clean -f
   ```

* **Resolving Conflicts:**
   - Manually handle conflicts that arise when merging branches.
   (Requires understanding conflict markers)

* **Stashing Changes:**
   - Temporarily save uncommitted changes for later use.
   ```bash
   git stash
   ```
   - Apply the most recent stashed changes.
   ```bash
   git stash pop
   ```

* **Ignoring Files:**
   - Configure Git to ignore specific files or patterns using a `.gitignore` file.

* **Branching Strategies:**
   - Utilize feature branches, hotfixes, and release branches for efficient workflow.

**Resources and Further Reading**

* **Git SCM - official documentation:** [https://git-scm.com/doc](https://git-scm.com/doc)
* **Interactive Git Tutorial:** [https://learngitbranching.js.org/](https://learngitbranching.js.org/)
* **Atlassian Git Tutorial:** [https://www.atlassian.com/git/tutorials](https://www.atlassian.com/git/tutorials)
* **Git Cheat Sheet:** [https://www.atlassian.com/git/tutorials/atlassian-git-cheatsheet](https://www.atlassian.com/git/tutorials/atlassian-git-cheatsheet)

**Note:** This guide provides a foundational understanding of Git commands. Refer to the resources above for in-depth explanations and advanced functionalities.
```typescript
import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

type Data = {
data: string
}

export default function LessonContent({ uri }: { uri: string }) {
	const [content, setContent] = useState('');

	useEffect(() => {
			async function fetchContent() {
			try {
			const response = await fetch(`/api/lessons?chapter=${uri}`, {
method: 'POST'
});
			const data: Data = await response.json() as Data;
			console.log(data)
			setContent(data.data);
			} catch (error) {
			console.error('Error fetching lesson content:', error);
			}
			}

			fetchContent().catch(console.error);
			}, [uri]);

return (
		<div className="h-full w-full">
		<ReactMarkdown>{content}</ReactMarkdown>
		</div>
       );
}
```
