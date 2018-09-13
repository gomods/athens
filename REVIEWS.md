# Pull Request Reviews

Whether you're contributing or reviewing pull requests, this document can help you!

- If you're reviewing, it can help you understand what to do
- If you're submitting a pull request, it can help you understand what folks
are going to do, how to help them, and how to get your pull request
merged faster

## Steps

We follow a lightweight list of steps in order to review 
[pull requests in the Athens project](https://github.com/gomods/athens/pulls)
(if you're not familiar with pull requests, see 
[here](https://help.github.com/articles/about-pull-requests/)).

Here they are:

- Anyone, whether or not they are an official [contributor](https://github.com/orgs/gomods/teams/contributors),
  can review a pull request.
- Pull requests must be reviewed by at least one 
  [maintainer](https://github.com/orgs/gomods/teams/maintainers), enforced by GitHub.
  Only maintainers can merge a pull request.
- For important changes we want to make sure that people in other time zones have a chance to
    review, so keep the pull request open for 24-36 hours before merging.
- Pull requests have to pass continuous integration (CI) tests.

# Review Types and When to Use Them

We use the [Github Pull Request Review](https://help.github.com/articles/about-pull-request-reviews/)
system to review pull requests. Overall, we think it's a pretty intuitive system
with a nice user interface. Hopefully you'll get the hang of it pretty quickly. If you
don't, no worries - ask us in the `#athens` room on the [Gophers Slack](https://invite.slack.golangbridge.org/)!

When you're doing a review on a PR, you'll make comments that nobody but you can see until you
submit the review. That feature can be nice because you might want to change things as you
learn more about the code, etc... 

It can also be confusing if you forget to submit your review! Lots of us have forgotten to do 
that :smile:.

Anyway, when you're done with your review and satisfied with all your comments, you'll have 
three options to submit your review:

- Comment
- Approve
- Request Changes

Below, we'll explain when to use each of these options.

## Request Changes

When you decide to submit your review, please use `Request Changes` when you've posted
comments asking the author to make some changes, explain something, etc..., and you 
want to block the PR from being merged until:

- The author has a chance to read your comments and make changes, etc...
- You have a chance to review again
- You explicitly approve in your review

## Comment

When you decide to submit your review, please use `Comment` for feedback that you'd
like addressed but it's not crucial if it's not. This review type will allow another maintainer
to merge the PR after your feedback has been addressed, without requiring you to 
come back and manually re-review and approve the PR. In other words, you won't
be blocking the PR from being merged.

## Approve

This one is the easiest. When you decide to submit your review, please use `Approve` if everything
looks good. If nobody else has any `Request Changes` or `Comment` reviews 
(GitHub will show a red "X" near their name if they do), you can click the "Squash and Merge"
button to merge their PR into the master branch!
