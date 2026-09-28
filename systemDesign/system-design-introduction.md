# System Design Introduction

## Introduction

Learn system design fast. All the essentials needed to pass a system design interview.

After conducting literally thousands of interviews at companies like Meta and Amazon, the most important things that candidates need to know to succeed in system design interviews have been collected. This approach takes a fundamentally different approach to teaching system design than you might find elsewhere, by working backwards from those things you need to know in order to succeed in an interview.

This is helpful for two reasons:
1. If you don't have a lot of time between now and your interview, you're going to learn the most impactful things as quickly as possible, and
2. As you learn new things you'll be able to connect them to real systems and real problems rather than just accumulating academic knowledge.

Other system design materials are either ChatGPT spew or go to a level of depth that you'll never possibly cover in an interview (and might be a yellow flag if you do). This guide is dense, practical, and efficient.

### What are system design interviews?

System design interviews are a way to assess your ability to take an ambiguously defined, high-level problem and break it down into the pieces of infrastructure that you'll need to solve it. These are *practical* interviews, not strictly academic ones, and most engineers find they are closer to real-world work than other types of interviews like the leetcode style interview.

Importantly, **design** interviews are not about getting to a single right answer. For many questions, there are many right answers. Instead your interviewer is looking to assess your ability to navigate a complex problem, reason about trade-offs, and communicate your thinking clearly.

Most entry-level software engineering roles will *not* have a system design interview (though there are plenty of exceptions). Once you've reached mid-level, system design interviews become more common. At the senior level, system design interviews are the norm and carry a disproportionate weight in the overall evaluation process for the candidate.

#### Types of System Design Interviews

Each company (and sometimes, each interviewer) will conduct a system design interview a little differently. The overwhelming majority of system design interviews will be what we'll call "Product Design" or "Infrastructure Design" interviews.

In these interviews you'll be asked to design a system behind a product or a system that supports a particular infrastructure use case like "Design a ride-sharing service like Uber" or "Design a rate limiter". These problems typically require infra: services, load balancers, databases, etc.

If you are planning for an interview where you'll be instead be asked to design the class structure of a system, that's an interview we call "Low-Level Design" (sometimes referred to as "Object-Oriented Design").

If your interview includes ML modelling, feature engineering, and other facets of an applied ML engineer's role, we call that "ML System Design".

Finally, if you're interviewing for a frontend engineering role, we highly recommend our friends at Great Frontend for both material and practice problems for frontend design interviews.

### Assessment

The interviewers conducting system design interviews are looking to assess certain skills and knowledge through the course of the interview.

At a high-level, while all candidates are expected to complete a full design satisfying the requirements, a mid-level engineer might cover the basics well but not into great depth, while a senior engineer will quickly work through the basics leaving time for them to show off the depth of their knowledge in deep dives.

Each company will have a different rubric for system design, but regardless of level these rubrics have strong themes that are common across all interviews: Problem Navigation, Solution Design, Technical Excellence, and Communication and Collaboration.

#### Problem Navigation

Your interviewer is looking to assess your ability to navigate a complex, under-specified problem. This means that you should be able to break down the problem into smaller, more manageable pieces, prioritize the most important ones, and then navigate through those pieces to a solution. This is often the most important part of the interview, and the part that most candidates (especially those new to system design) struggle with.

The most common ways that candidates fail with this competency are:
- Insufficiently exploring the problem and gathering requirements.
- Focusing on uninteresting/trivial aspects of the problem vs the most important ones.
- Getting stuck on a particular piece of the problem and not being able to move forward.
- Failing to deliver a working system.

The reason many candidates fail to make progress in their interview is due to a lack of structure in their approach.

#### Solution Design

With a problem broken down, your interviewer wants to see how you can solve each of the pieces of the problem. This is where your knowledge of the Core Concepts comes into play. You should be able to describe how you would solve each piece of the problem, and how those pieces fit together into a cohesive whole.

The most common ways that candidates fail with this competency are:
- Not having a strong enough understanding of the core concepts to solve the problem.
- Ignoring scaling and performance considerations.
- "Spaghetti design" - a solution that is not well-structured and difficult to understand.

Interviewers are on alert for candidates who have simply memorized answers or material. They'll test you by probing your reasoning, doubting your answers, or asking you to explore tradeoffs.

#### Technical Excellence

Your interviewer is looking to assess your technical depth and ability to make sound technical decisions. This includes your understanding of key technologies, your ability to reason about trade-offs, and your knowledge of best practices.

The most common ways that candidates fail with this competency are:
- Not having a strong enough understanding of the key technologies to make informed decisions.
- Making decisions without considering trade-offs.
- Not being able to justify your technical choices.

#### Communication and Collaboration

Your interviewer is looking to assess your ability to communicate your thinking clearly and collaborate effectively. This includes your ability to explain your design decisions, respond to feedback, and work with the interviewer to refine your solution.

The most common ways that candidates fail with this competency are:
- Not communicating your thinking clearly.
- Not being responsive to feedback.
- Not being able to explain your design decisions.

## How to Prepare for System Design Interviews

Having helped thousands of candidates pass their FAANG interviews, we've learned exactly what works.

### Choose how you want to prepare

We recommend the course if you want a clear order to follow. If you already know what you need, use this page as a plan and choose the lessons and practice problems yourself.

#### Build a Foundation

1. **Understand what a system design interview is**: Maybe you've never done a system design interview before, you're not alone! Start by reading this intro to system design or watching a video of a mock system design interview.

2. **Choose a delivery framework**: System design interviews move fast. It's important that you have a clear roadmap to help you think linearly and avoid scope creep. Use a Delivery Framework. This is the framework you'll follow to design your system come interview day.

3. **Start with the basics**: If you're new to system design in particular, you'll want to start by learning the basics and mapping out the scope of knowledge required. Start by reading about the Core Concepts, Key Technologies, and Common Patterns used in system design interviews.

#### Practice Practice Practice

Once you have the foundation in place, it's time to practice. Passively consuming content is good, but you'll retain 10x more information by actually doing.

1. **Choose a question**: Select a question from the list of common questions.

2. **Read the requirements**: Understand the requirements of the system you'll need to design.

3. **Try to answer on your own**: Either practice with Guided Practices or on a virtual whiteboard like Excalidraw.

4. **Read the answer key**: Only after you have tried to answer the question, read the answer key to see how your answer compares.

5. **Put your knowledge to the test**: Once you've done a few questions and are feeling comfortable, run a peer mock with someone who works at your target company — telling your design out loud under time pressure is a different skill than reading about it.

### Common System Design Questions

Here are some common system design interview questions you might encounter:

**Easy Questions:**
- Bitly
- Dropbox
- Yelp
- Local Delivery Service

**Medium Questions:**
- Ticketmaster
- Instagram
- FB News Feed
- Tinder
- LeetCode
- WhatsApp
- Strava
- Distributed Cache
- Rate Limiter
- Online Auction
- YouTube
- Job Scheduler
- FB Live Comments
- YouTube Top K
- Uber
- Web Crawler
- Ad Click Aggregator
- FB Post Search
- News Aggregator
- Price Tracking Service
- Notification System
- Robinhood
- Google Docs
- Payment System
- Metrics Monitoring
- Online Chess
- ChatGPT
- Flash Sale

---

*This is Part 1. Part 2 covers System Design Core Concepts.*