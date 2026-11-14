## History

Habbowidgets has been live since 2015, most of the code for the actual widgets came from the Widgets section of Habbies.nl, the fansite I was operating before that.

It hasn't gotten any big features since 2019. For me personally, I too started to visit Habbo less and less. My first account just celebrated 25 years, which is honestly quite an achievement in internet years, so kudos Sulake!

I still like communities and programming, so I'll probably keep Habbowidgets up and running. However last year we really started to hit some technical limits and this year Habbo has been stricter with regards to the APIs, we had some downtime.

## Technical details

The website has been running for well over 10 years, and we've had well over 2 million accounts stored, but we were also running into disk space issues.

So last year I started to add some clean up tasks, badges, friends, groups that had been removed for over 6 months, were going to be deleted from the database.
Habbos who haven't had any updates for a year since their last check, were going to be removed as well.

<figure class="center">
    <img src="/kd/images/blog/2026/habbowidgets-stats.png" alt="The amount of Habbos and unique badges per hotel; 200k-300k for .com, .es and .com.br, while the rest has between 20k and 50k. Badge distribution is more uniform, between 10k and 15k." class="u-photo">
    <figcaption>The amount of Habbos and unique badges per hotel.</figcaption>
</figure>

Right now we still have over a million Habbos, and those are just the active ones. Our disk, mostly our database data, is at 95GB of 100.
Part of this is also because of some inefficient table structures, specifically for the friends table.

Based on the above table, it might also make sense to partition our database for com.br, es, com and another section for 'Europe'.

The last change I made was making some parts of the Habinfo page a bit more mobile friendly. Habbowidgets was always made with a responsive layout in mind,
but personally I always view it from a desktop. In the early years it was also 70-30 split between desktop and mobile, but in recent years it's become 50-50, if not even more towards mobile.

The other change I made this year, was adding some functionality to add a limiter to the API. Habbo only allows us to request 1 call per second, which is fair.

While not a change I made, this year I've also been fighting AI bots a lot. It's causing a lot of load, and specifically when they hit the Habinfo pages and use our already limited requests for the API calls. Two years ago most bots would still announce themselves, but since 2024 they're really becoming a pest, using botnets (hacked PCs) to mimic regular users, using normal browsers and IPs that look like regular customers on internet providers.

## Questioning

Without a doubt Habbowidgets is gonna stay alive, but it might need a fresh make-over.

I want to freshen the layout, but only slightly. I think it needs even more of a mobile-first approach.
One of the great things we had, was that we showed all badges at once, but of course this was when people had like 400 to 600 badges.
The top 250 now has people with between 5000 and 9000 badges.

So we probably will have to add pagination to show it per 1000, often times people are only interested in the latest newest badges, so that should be fine.

The same pagination idea has to happen for Friends as well, maybe even show only 500 at a time. Loading these badges and friends also adds a lot of requests to Habbo's servers, so I'd also like to be cautious of that.

Something else I've been debating is whether to have user accounts, so you can sign up and login and connect your Habbo account. I'm not sure what extra benefit that would give.

Internally I'm going to set up several queues that we'll use to update Habbos, there will be different priorities, so perhaps if you're logged in you can update with highest priority and any visitor not logged in (or bot) goes in a slower queue.
