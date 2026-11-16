# Contributing to Kindling
Thank you for your interest in contributing to Kindling! We welcome all people who want to contribute in a healthy and constructive manner within our community. 
​

## Become a contributor
You can contribute to Kindling in several ways, like:
- Contribute to the codebase
- Report and triage bugs
- Write technical documentation and blog posts for users and contributors
- Help others by answering questions about Kindling

For more ways to contribute, check out the [Open Source Guides](https://opensource.guide/how-to-contribute/).
​

## Report bugs
Before submitting a new issue, try to make sure someone hasn’t already reported the problem. You can look through the existing issues for similar issues.
​

[Report a bug](https://github.com/KindlingProject/kindling/issues/new?assignees=&labels=&template=bug_report.md&title=) by submitting a new issue. Make sure to follow the issue template and add more detailed information which will help us reproduce the problem.

## Triage issues
If you don't have the knowledge or time to code, consider helping with issue triage. The community will thank you for saving them time by spending some of yours.
Read more about the ways you can [Triage issues](contribute/triage_issues.md).

## Answering questions
It's important to us to help the users who have problems with Kindling, and we’d love your help. Go to [disscussions](https://github.com/KindlingProject/kindling/discussions), you can find unanswered questions, and you can answer those questions.


## Your first contribution
The first step to starting to contribute to Kindling is finding something to work on. You can start by fixing beginner-friendly issues or improving Kindling documents, no contribution is too small!

+ How to find beginner-friendly issues? Kindling has a `good first issue` label for issues that don’t need high-level knowledge to contribute. You can browse issues labeled with `good first time`. 

+ How to find documents improving issues? Kindling has a `documentation` label for issues that you can improve Kindling docs.

If you are ready to contribute code changes, review the [developer guide](http://kindling.harmonycloud.cn/docs/developer-guide/build-kindling-container-image/) for how to set up your local environment.
When you want to submit your local changes, read about [create pull request](contribute/create_pull_request.md).

## Sign your commits

The sign-off is a simple line at the end of the explanation for a commit. All commits need to be signed. Your signature certifies that you wrote the patch or otherwise have the right to contribute the material. The rules are pretty simple, if you can certify the rules (from [developercertificate.org](https://developercertificate.org/)), then you just need to add a line to every git commit message, like:

```
Signed-off-by: lina <lina@example.cn>
```

### Configure auto commit signing in git 

1. Set your `user.name` and `user.email` using the following commands:

   ```
   git config --global user.name lina
   
   git config --global user.email lina@example.cn
   ```

2. Sign your commit with `git commit -s`, then use `git log` to verify that the signed-off message is added.

   ```
   Author: lina <lina@example.cn>
   
   Date:  Tue Jan 25 09:42:40 2022 +0800
   
     add content
   
     Signed-off-by: lina <lina@example.cn>
   ```

   

