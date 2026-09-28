# How Figma Built Multiplayer Editing on Simplified CRDTs

Dealing with Contention, Real-time Updates

## The TLDR

When Figma launched multiplayer editing in September 2016, no other design tool had live collaboration. In fact, most designers hated the idea of it, fearing "hovering art directors" and "design by committee" (funny to think about now). They carried on anyway, and three years later Evan Wallace, Figma's co-founder, published a walkthrough of how the syncing actually works for others to learn from.

The system's less complex than you'd guess. To handle collaborative editing, they considered operational transformation (OT), the algorithm behind Google Docs, but decided it was overkill and would be incredibly difficult to get right for a design tool. They studied CRDTs (conflict-free replicated data types) and found that most of the complexity exists to satisfy a decentralized environment, which wasn't something they'd need so long as every edit to a document flowed through a single server. What was left after cutting both was a document modeled as a tree of objects, a single conflict rule where the last write to reach the server wins when two people write the same property of the same object, and a central server that steps in only to ensure the document remains a valid tree.

Figma's solution aged into the industry's answer. The design they derived from the research, a central server ordering optimistic edits with last-writer-wins per property, is what teams now buy off the shelf as a sync engine.

## The Problem

### One document, many editors

For those unfamiliar with Figma, it's a design tool that runs in the browser and is used by product teams to draw the screens and icons that become their apps. Figma's goal was to allow multiple designers to edit the same document at the same time.

The challenge is that every editor's copy of the document must converge to a single source of truth. When the edits stop, everyone has to be looking at an identical version of the document, just like with Google Docs.

Edits also have to feel instant. Nobody wants to drag a shape and wait on a central server round trip before it moves on the screen, so each client applies its own changes immediately and has to sync them afterward. This creates a short window where the two copies really do disagree, until the sync catches up.

Users also need to be able to make edits offline, which stretches that "out of sync" window from seconds to hours. A designer on a plane keeps working, then reconnects and merges back in, so the system has to absorb whole batches of stale edits.

Lastly, to add to the complexity, a Figma file is a tree. Frames hold shapes which hold text, etc, and two people restructuring it at once can do things no single-player editor could, like moving the same object into two different frames at the same time. The sync system has to figure out how to handle those moves without ever duplicating an object or losing part of the document.

## The Solution

### Simplified CRDTs

At the time, the field offered two potential answers to this multiplayer problem: OT and CRDTs.

Figma ruled OT out quickly. It works by expressing every edit as an operation and transforming operations against concurrent ones, so an insert at position 5 becomes an insert at position 8 when someone else added three characters ahead of it. It's famously hard to get right.

CRDTs looked closer to what they needed. A CRDT solves convergence by letting each replica accept edits entirely on its own, with merges designed to give the same result no matter what order the edits happened in. The simplest example is a grow-only set, where the only operation is adding and merging two copies is just taking the union, so it can't matter whose edit came first or how long the copies were apart.

A real document needs more than add-only, though. The moment values can change, someone has to say which of two competing writes came second, and the CRDT answer is the last-writer-wins register. It holds a single value, every write carries a timestamp, and merging keeps whichever value bears the newer one.

But keeping track of all that metadata in a register has real cost and complexity. The bookkeeping piles up from there. Deletes have to leave markers behind for replicas that come back late instead of actually removing anything, all of it more metadata answering the same question without anyone in charge, which write came last.

Figma, though, has exactly the thing CRDTs are designed to live without, a central server. Every edit already flows through it, so the order can simply be determined by the server, which removes the timestamps and most of the bookkeeping. The register's rule survives, the last write wins, with last now meaning last to reach the server. The team went through the rest of the CRDT literature the same way, cutting every piece a central server makes unnecessary. What survived the cutting is a document whose every property is one of these registers, minus the timestamp. A simplified CRDT.

### Last writer wins on each property

Two people change the same rectangle at the same moment. Which edit wins?

Figma's tree of objects works very much like the HTML DOM. You have a single root object representing the document, whose children are the pages, and under each page sits the hierarchy of frames, shapes, and text you see in the layers panel. Each object in this tree has its own ID and a set of properties with values, which means the whole document fits in one structure:

```
Map<ObjectID, Map<Property, Value>>
```

The server keeps, for every object and property, the latest value any client has sent. If you change the color of a rectangle while a teammate resizes it, then both edits land, because they touch different properties. Change the color of the same rectangle at the same moment and there's a conflict. But it can be resolved by simply keeping whichever value reached the server last. Each property is the last-writer-wins register from earlier.

The property value is the unit of atomicity. If a text property holds B and one client changes it to AB while another concurrently changes it to BC, the document ends up as AB or BC, never a merged ABC. For a text editor, that would be disqualifying, but it makes perfect sense for visual objects, where an edit sets a whole value like a fill color or a width. When two people recolor the same shape at the same moment, one color winning outright is the outcome you'd actually want, since there's no sensible way to merge red and blue anyway.

### The server's role in maintaining validity

The last-writer-wins rule handles property conflicts, but it doesn't prevent structural conflicts like moving the same object into two different frames at once. That's where the server steps in.

The server validates every edit to ensure the document remains a valid tree. If two clients try to move the same object, the server rejects the second move and sends an error back to the client. The client then refreshes its state and tries again.

This is a simple but effective approach. The server doesn't need complex conflict resolution for structural changes - it just enforces validity and lets clients retry when conflicts occur.

## Key Takeaways

1. **Central servers simplify things**: When you have a central server, you can dramatically simplify CRDTs by removing the need for distributed coordination.

2. **Last-writer-wins per property**: For visual objects, last-writer-wins on each property is often sufficient and simpler than trying to merge values.

3. **Separate structural from property conflicts**: Handle property conflicts with simple rules, and structural conflicts with server validation.

4. **Optimistic updates with conflict resolution**: Clients apply changes immediately, then sync with the server, retrying if conflicts occur.