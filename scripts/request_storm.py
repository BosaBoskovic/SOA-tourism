#!/usr/bin/env python3
"""Populates SOA-tourism with realistic data and fires concurrent request
storms at it through the gateway - for demoing the monitoring stack (Grafana/
Prometheus/Loki/Jaeger/RabbitMQ) under real traffic, live.

Safe to run over and over: every account this script creates gets a fresh,
timestamped username, so nothing collides with a previous run. A few actions
are genuinely one-time per (user, thing) by business rule - reviewing the
same tour twice, re-publishing an already-published tour, registering the
same username twice - those are recognised by message and reported
separately as "expected rejections", not counted as bugs.

Usage:
    python scripts/request_storm.py seed                     # populate data
    python scripts/request_storm.py seed --guides 8 --tourists 24
    python scripts/request_storm.py storm                    # generic load burst
    python scripts/request_storm.py storm --requests 5000 --concurrency 100
    python scripts/request_storm.py checkout-stress          # hammer checkout specifically
    python scripts/request_storm.py throttle-demo            # show off the login backoff
    python scripts/request_storm.py all                      # seed, then storm

State (accounts/tours) is kept in scripts/.storm_state.json between runs so
storm/checkout-stress/throttle-demo can reuse whatever seed has already
created without you having to re-seed every time.
"""

import argparse
import json
import os
import random
import sys
import time
import urllib.request
import urllib.error
from concurrent.futures import ThreadPoolExecutor, as_completed

BASE_URL = os.environ.get("STORM_BASE_URL", "http://localhost:8080")
STATE_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".storm_state.json")
RUN_ID = f"{int(time.time())}{random.randint(100, 999)}"

# Substrings of known, expected "you already did that" business-rule errors -
# matching one of these means "working as intended", not "found a bug".
EXPECTED_REJECTIONS = [
    "vec postoje", "already exists", "already reviewed", "u korpi",
    "already in", "Morate zapratiti", "must follow", "tour is not in draft",
    "korpa je prazna", "cart is empty", "already following",
]


# --------------------------------------------------------------- plumbing --

class Stats:
    def __init__(self):
        self.ok = 0
        self.expected_rejections = []
        self.unexpected = []
        self.latencies = []

    def record(self, label, status, body, expected_statuses, timeout_s=None):
        latency = timeout_s
        if latency is not None:
            self.latencies.append(latency)
        if status in expected_statuses:
            self.ok += 1
            return True
        text = json.dumps(body) if not isinstance(body, str) else body
        if any(marker.lower() in text.lower() for marker in EXPECTED_REJECTIONS):
            self.expected_rejections.append((label, status, text[:150]))
            return False
        self.unexpected.append((label, status, text[:200]))
        return False

    def print_summary(self, title):
        print()
        print("=" * 64)
        print(title)
        print("=" * 64)
        print(f"ok:                  {self.ok}")
        print(f"expected rejections: {len(self.expected_rejections)}  (business rules working as intended)")
        print(f"unexpected errors:   {len(self.unexpected)}")
        if self.latencies:
            lat = sorted(self.latencies)
            p50 = lat[int(len(lat) * 0.50)] * 1000
            p90 = lat[min(len(lat) - 1, int(len(lat) * 0.90))] * 1000
            p99 = lat[min(len(lat) - 1, int(len(lat) * 0.99))] * 1000
            print(f"latency p50/p90/p99: {p50:.0f}ms / {p90:.0f}ms / {p99:.0f}ms")
        if self.expected_rejections and len(self.expected_rejections) <= 30:
            print("\nexpected rejections (business rules doing their job):")
            for label, status, body in self.expected_rejections:
                print(f"  - {label}: {status} {body}")
        if self.unexpected:
            print("\nunexpected errors (worth a closer look):")
            for label, status, body in self.unexpected[:20]:
                print(f"  ! {label}: {status} {body}")
            if len(self.unexpected) > 20:
                print(f"  ... and {len(self.unexpected) - 20} more")


def call(method, path, body=None, token=None, timeout=15):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE_URL + path, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    t0 = time.monotonic()
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read()
            return resp.status, (json.loads(raw) if raw else {}), time.monotonic() - t0
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            parsed = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            parsed = {"raw": raw.decode(errors="replace")}
        return e.code, parsed, time.monotonic() - t0
    except Exception as e:  # noqa: broad - a storm must survive one bad request
        return -1, {"exception": str(e)}, time.monotonic() - t0


def load_state():
    if os.path.exists(STATE_PATH):
        with open(STATE_PATH) as f:
            return json.load(f)
    return {"guides": {}, "tourists": {}, "published_tours": [], "following": {}}


def save_state(state):
    with open(STATE_PATH, "w") as f:
        json.dump(state, f)


# ------------------------------------------------------------------ seed --

LANDMARKS = [
    ("Petrovaradinska tvrdjava", 45.2551, 19.8681),
    ("Dunavski park", 45.2559, 19.8399),
    ("Zmaj Jovina ulica", 45.2547, 19.8452),
    ("Trg slobode", 45.2551, 19.8452),
    ("Strand", 45.2469, 19.8517),
    ("Sinagoga", 45.2560, 19.8464),
    ("Muzej Vojvodine", 45.2536, 19.8442),
    ("Liman promenada", 45.2413, 19.8390),
]
TAGS_POOL = ["priroda", "istorija", "hrana", "avantura", "porodica", "arhitektura", "kultura"]
DIFFICULTIES = ["easy", "medium", "hard"]
TOUR_NOUNS = ["Setnja", "Obilazak", "Avantura", "Prica o", "Tajne"]
TOUR_TOPICS = ["Petrovaradina", "starog grada", "Dunava", "tvrdjave", "centra grada"]
BLOG_TITLES = ["Moj vikend u Novom Sadu", "5 mesta koja morate posetiti",
               "Kako izgleda prava avantura", "Utisci sa poslednje ture", "Vodic za pocetnike"]
COMMENT_TEXTS = ["Super post!", "Hvala na preporuci.", "I ja sam bio tamo, slazem se.", "Odlicne fotke."]
REVIEW_COMMENTS = [
    "Odlicno iskustvo, vodic je bio sjajan!", "Prelepa ruta, preporucujem svima.",
    "Malo naporno za pocetnike ali vredi truda.", "Fantasticni pogledi, ponovicu ovu turu.",
    "Solidna tura, moglo je i bolje sa vodjenjem.",
]
GUIDE_NAMES = ["nikolina", "marko", "jovana", "stefan", "dragana", "milos", "jasmina", "vuk"]
TOURIST_NAMES = [
    "ana", "petar", "milica", "luka", "sara", "aleksandar", "teodora", "nemanja",
    "ivana", "filip", "marija", "uros", "jelena", "vladimir", "tijana", "dusan",
    "katarina", "bojan", "milena", "igor", "sandra", "nikola", "ljiljana", "darko",
]


def cmd_seed(args):
    stats = Stats()
    state = load_state()
    guides, tourists = {}, {}

    print(f"== registering {args.guides} guides + {args.tourists} tourists (run id {RUN_ID}) ==")
    for name in random.sample(GUIDE_NAMES, k=min(args.guides, len(GUIDE_NAMES))):
        username = f"{name}-{RUN_ID}"
        status, body, dt = call("POST", "/stakeholders/register", {
            "username": username, "password": "CorrectHorse1!",
            "email": f"{username}@example.com", "role": "guide",
        })
        if stats.record(f"register guide {username}", status, body, (201,), dt):
            guides[username] = body["accessToken"]
    for name in random.sample(TOURIST_NAMES, k=min(args.tourists, len(TOURIST_NAMES))):
        username = f"{name}-{RUN_ID}"
        status, body, dt = call("POST", "/stakeholders/register", {
            "username": username, "password": "CorrectHorse1!",
            "email": f"{username}@example.com", "role": "tourist",
        })
        if stats.record(f"register tourist {username}", status, body, (201,), dt):
            tourists[username] = body["accessToken"]
    print(f"  {len(guides)} guides, {len(tourists)} tourists ready")

    print("== creating + publishing tours (with keypoints) ==")
    published_tours = []
    for gi, (guide, token) in enumerate(guides.items()):
        for t in range(random.randint(2, 3)):
            name = f"{random.choice(TOUR_NOUNS)} {random.choice(TOUR_TOPICS)} #{t+1}"
            difficulty = DIFFICULTIES[(gi + t) % len(DIFFICULTIES)]
            tags = random.sample(TAGS_POOL, k=random.randint(2, 4))
            status, body, dt = call("POST", "/tours", {
                "name": name,
                "description": f"Nezaboravna tura kroz Novi Sad - {name}, za sve ljubitelje {tags[0]}.",
                "difficulty": difficulty, "tags": tags,
                "durations": [{"transport": "walk", "minutes": random.choice([45, 60, 90, 120])}],
            }, token=token)
            if not stats.record(f"create tour ({guide})", status, body, (201,), dt):
                continue
            tour_id = body["id"]

            chosen = random.sample(LANDMARKS, k=random.randint(2, 4))
            last_keypoint_id = None
            for order, (lname, lat, lng) in enumerate(chosen):
                status, kbody, dt = call("POST", "/keypoints", {
                    "tourId": tour_id, "name": lname, "description": f"Stanica {order+1}: {lname}",
                    "latitude": lat, "longitude": lng, "imageUrl": "", "order": order,
                    "lengthKm": 0.5 if order > 0 else None,
                }, token=token)
                if stats.record(f"add keypoint ({tour_id})", status, kbody, (201,), dt):
                    last_keypoint_id = kbody.get("id")

            price = random.choice([500, 800, 1200, 1500, 2000, 2500])
            status, ubody, dt = call("PUT", f"/tours/{tour_id}", {
                "name": name, "description": f"Nezaboravna tura kroz Novi Sad - {name}.",
                "difficulty": difficulty, "tags": tags,
                "durations": [{"transport": "walk", "minutes": 60}], "price": price,
            }, token=token)
            stats.record(f"set price ({tour_id})", status, ubody, (200,), dt)

            status, pbody, dt = call("PUT", f"/tours/{tour_id}/publish", token=token)
            if stats.record(f"publish ({tour_id})", status, pbody, (200,), dt):
                published_tours.append({"id": tour_id, "price": price, "author": guide})

            # a guide-created "encounter" challenge tied to the last keypoint - optional flavor
            if last_keypoint_id and random.random() < 0.5:
                lname, lat, lng = chosen[-1]
                status, ebody, dt = call("POST", "/encounters", {
                    "name": f"Pronadji {lname}", "description": "Priblizi se lokaciji da preuzmes nagradu.",
                    "tourId": tour_id, "keyPointId": last_keypoint_id,
                    "latitude": lat, "longitude": lng, "radiusMeters": 50,
                    "reward": random.choice(["Besplatna razglednica", "10% popusta na sledecu turu", "Bedz osvajaca"]),
                }, token=token)
                stats.record(f"create encounter ({tour_id})", status, ebody, (201,), dt)
    print(f"  {len(published_tours)} tours published")

    print("== building the social graph (follows -> notification-delivery) ==")
    all_usernames = list(guides.keys()) + list(tourists.keys())
    following = {u: set() for u in all_usernames}
    follow_count = 0
    for follower, token in list(tourists.items()) + list(guides.items()):
        # everyone follows every guide, so blog posts are always commentable
        targets = set(guides.keys()) - {follower}
        targets |= set(random.sample([u for u in all_usernames if u != follower],
                                      k=min(3, max(0, len(all_usernames) - 1))))
        for target in targets:
            status, body, dt = call("POST", "/followers/follow", {"targetUsername": target}, token=token)
            if stats.record(f"follow ({follower}->{target})", status, body, (201,), dt):
                follow_count += 1
                following[follower].add(target)
    # a few unfollows too, for variety and to exercise that path
    unfollow_count = 0
    for follower in random.sample(all_usernames, k=max(1, len(all_usernames) // 5)):
        if not following[follower]:
            continue
        target = random.choice(list(following[follower]))
        token = guides.get(follower) or tourists.get(follower)
        status, body, dt = call("DELETE", f"/followers/follow/{target}", token=token)
        if stats.record(f"unfollow ({follower}->{target})", status, body, (200, 204), dt):
            unfollow_count += 1
            following[follower].discard(target)
    print(f"  {follow_count} follows, {unfollow_count} unfollows")

    print("== shopping carts + checkouts (purchase-completed) ==")
    purchases = []
    for tourist, token in tourists.items():
        if not published_tours or random.random() < 0.2:
            continue
        picks = random.sample(published_tours, k=min(len(published_tours), random.randint(1, 2)))
        for tour in picks:
            status, body, dt = call("POST", f"/shopping-cart/{tourist}/items", {"tourId": tour["id"]}, token=token)
            stats.record(f"add to cart ({tourist})", status, body, (200, 201), dt)
        status, body, dt = call("POST", f"/checkout/{tourist}", token=token)
        if stats.record(f"checkout ({tourist})", status, body, (200,), dt) and isinstance(body, list):
            for row in body:
                purchases.append((tourist, row["tourId"]))
    print(f"  {len(purchases)} purchases completed")

    print("== reviews on purchased tours ==")
    review_count = 0
    for tourist, tour_id in purchases:
        token = tourists[tourist]
        status, body, dt = call("POST", "/reviews", {
            "tourId": tour_id, "touristId": tourist, "touristName": tourist,
            "rating": random.randint(3, 5), "comment": random.choice(REVIEW_COMMENTS),
            "images": [], "tourVisitDate": "2026-09-01",
        }, token=token)
        if stats.record(f"review ({tourist}->{tour_id})", status, body, (201,), dt):
            review_count += 1
    print(f"  {review_count} reviews created")

    print("== simulating tourists moving during a tour (tourist-position) ==")
    position_count = 0
    for tourist, tour_id in purchases[: max(1, len(purchases) // 2)]:
        token = tourists[tourist]
        lat, lng = random.choice(LANDMARKS)[1:]
        for _ in range(random.randint(1, 3)):
            lat += random.uniform(-0.001, 0.001)
            lng += random.uniform(-0.001, 0.001)
            status, body, dt = call("PUT", "/tourist-position", {
                "touristId": tourist, "latitude": lat, "longitude": lng,
            }, token=token)
            if stats.record(f"position update ({tourist})", status, body, (200, 201), dt):
                position_count += 1
    print(f"  {position_count} position updates")

    print("== blog posts + comments + likes ==")
    blog_ids_by_author = {}
    for username, token in guides.items():
        for _ in range(random.randint(1, 2)):
            title = f"{random.choice(BLOG_TITLES)} - {username}"
            status, body, dt = call("POST", "/blog", {
                "title": title,
                "descriptionMarkdown": f"# {title}\n\nOvo je moja prica o poslednjoj turi kroz grad.",
                "imageUrls": [],
            }, token=token)
            if stats.record(f"create blog ({username})", status, body, (201,), dt):
                blog_ids_by_author.setdefault(username, []).append(body["blog"]["id"])

    comment_count, like_count, edit_count = 0, 0, 0
    everyone = list(guides.items()) + list(tourists.items())
    blog_ids = [bid for ids in blog_ids_by_author.values() for bid in ids]
    for author, ids in blog_ids_by_author.items():
        commenters = random.sample([(u, tok) for u, tok in everyone if u != author],
                                    k=min(3, max(1, len(everyone) - 1)))
        for blog_id in ids:
            for username, token in commenters:
                status, body, dt = call("POST", f"/blog/{blog_id}/comments", {"text": random.choice(COMMENT_TEXTS)}, token=token)
                if stats.record(f"comment ({username}->{blog_id})", status, body, (201,), dt):
                    comment_count += 1
                    comment_id = ((body.get("blog") or {}).get("comments") or [{}])[-1].get("id")
                    if comment_id and random.random() < 0.3:
                        status, ebody, dt = call("PUT", f"/blog/{blog_id}/comments/{comment_id}",
                                                  {"text": random.choice(COMMENT_TEXTS) + " (izmenjeno)"}, token=token)
                        if stats.record(f"edit comment ({blog_id})", status, ebody, (200,), dt):
                            edit_count += 1
                status, body, dt = call("POST", f"/blog/{blog_id}/like", token=token)
                if stats.record(f"like ({username}->{blog_id})", status, body, (200,), dt):
                    like_count += 1
    print(f"  {len(blog_ids)} posts, {comment_count} comments ({edit_count} edited), {like_count} likes")

    # merge into persistent state so storm/checkout-stress/throttle-demo can reuse it
    state["guides"].update(guides)
    state["tourists"].update(tourists)
    state["published_tours"].extend(published_tours)
    save_state(state)

    stats.print_summary("SEED SUMMARY")
    print(f"\nTotal in {STATE_PATH}: {len(state['guides'])} guides, "
          f"{len(state['tourists'])} tourists, {len(state['published_tours'])} tours")


# ----------------------------------------------------------------- storm --

def cmd_storm(args):
    state = load_state()
    guides, tourists, tours = state["guides"], state["tourists"], state["published_tours"]
    if not guides and not tourists:
        print("No seed data found - run 'seed' first (or pass --requests-only-reads to skip needing accounts).")
        sys.exit(1)
    all_tokens = list(guides.values()) + list(tourists.values())
    tourist_pairs = list(tourists.items())
    all_usernames = list(guides.keys()) + list(tourists.keys())
    print(f"Using {len(guides)} guides, {len(tourists)} tourists, {len(tours)} tours from {STATE_PATH}")

    def one_request():
        r = random.random()
        if r < 0.30:
            return call("GET", "/tours")
        if r < 0.40:
            diff = random.choice(DIFFICULTIES)
            return call("GET", f"/tours?difficulty={diff}&sortBy=price")
        if r < 0.55:
            if tours:
                t = random.choice(tours)
                return call("GET", f"/tours/{t['id']}")
            return call("GET", "/tours")
        if r < 0.62:
            return call("GET", "/health")
        if r < 0.70 and tourist_pairs:
            username, token = random.choice(tourist_pairs)
            return call("GET", "/stakeholders/profile", token=token)
        if r < 0.78:
            token = random.choice(all_tokens)
            return call("GET", "/blog", token=token)
        if r < 0.85 and tours:
            t = random.choice(tours)
            return call("GET", f"/reviews/tour/{t['id']}")
        if r < 0.92 and tourist_pairs:
            username, token = random.choice(tourist_pairs)
            if tours:
                t = random.choice(tours)
                return call("POST", f"/shopping-cart/{username}/items", {"tourId": t["id"]}, token=token)
        # remaining: more follows, keeps RabbitMQ traffic flowing too
        follower = random.choice(all_usernames)
        target = random.choice([u for u in all_usernames if u != follower])
        token = guides.get(follower) or tourists.get(follower)
        return call("POST", "/followers/follow", {"targetUsername": target}, token=token)

    _run_burst(one_request, args.requests, args.concurrency, "GENERIC STORM")


def cmd_checkout_stress(args):
    """Hammers checkout specifically: cross-service (gateway -> payments gRPC
    -> Postgres -> RabbitMQ purchase-completed -> tours consumer) in one
    concentrated burst, instead of the generic storm's mixed traffic."""
    state = load_state()
    tourists, tours = state["tourists"], state["published_tours"]
    if not tourists or not tours:
        print("Need seeded tourists + published tours - run 'seed' first.")
        sys.exit(1)
    print(f"Using {len(tourists)} tourists, {len(tours)} tours from {STATE_PATH}")

    # give every tourist something in their cart first (best-effort - some
    # will legitimately fail with "already in cart" on a rerun, that's fine)
    print("Priming shopping carts...")
    for username, token in tourists.items():
        t = random.choice(tours)
        call("POST", f"/shopping-cart/{username}/items", {"tourId": t["id"]}, token=token)

    tourist_pairs = list(tourists.items())

    def one_checkout():
        username, token = random.choice(tourist_pairs)
        return call("POST", f"/checkout/{username}", token=token)

    _run_burst(one_checkout, args.requests, args.concurrency, "CHECKOUT STRESS")


def _run_burst(request_fn, total, concurrency, title):
    print(f"\nFiring {total} requests at concurrency {concurrency}...")
    stats = Stats()
    start = time.monotonic()
    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(request_fn) for _ in range(total)]
        done = 0
        for fut in as_completed(futures):
            status, body, dt = fut.result()
            stats.record("burst request", status, body, (200, 201, 204), dt)
            done += 1
            if done % max(1, total // 10) == 0:
                print(f"  {done}/{total} done...")
    elapsed = time.monotonic() - start
    print(f"\nwall time: {elapsed:.1f}s   throughput: {total/elapsed:.1f} req/s")
    stats.print_summary(title)


# ------------------------------------------------------------ throttling --

def cmd_throttle_demo(args):
    """Registers one dedicated account, then hammers /stakeholders/login
    with a wrong password to show the progressive backoff live: first 2
    failures free, then 1s, 2s, 4s, 8s, ... - a good thing to have on
    screen for a demo since it's a custom feature, not off-the-shelf."""
    username = f"throttle-demo-{RUN_ID}"
    status, body, _ = call("POST", "/stakeholders/register", {
        "username": username, "password": "CorrectHorse1!",
        "email": f"{username}@example.com", "role": "tourist",
    })
    if status != 201:
        print(f"Could not register demo account: {status} {body}")
        sys.exit(1)
    print(f"Demo account: {username}\n")
    print(f"{'attempt':>8} {'status':>7} {'retry-after':>12}   body")
    print("-" * 70)

    attempts = args.attempts
    for i in range(1, attempts + 1):
        status, body, _ = call("POST", "/stakeholders/login", {
            "usernameOrEmail": username, "password": "definitely-wrong-password",
        })
        retry_after = body.get("retryAfterSeconds") if isinstance(body, dict) else None
        print(f"{i:>8} {status:>7} {str(retry_after or '-'):>12}   {json.dumps(body)}")
        if retry_after:
            wait = retry_after + 0.2
            print(f"         (waiting {wait:.1f}s for the lockout to expire...)")
            time.sleep(wait)
    print("\nDone. Go check the Loki 'level=warn' logs for stakeholders, and notice "
          "the retryAfterSeconds doubling each time a real attempt gets through.")


# ---------------------------------------------------------------- cli ----

def main():
    global BASE_URL
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--base-url", default=None, help=f"Gateway URL (default: {BASE_URL})")
    sub = parser.add_subparsers(dest="command", required=True)

    p_seed = sub.add_parser("seed", help="Populate accounts/tours/follows/purchases/reviews/blog data")
    p_seed.add_argument("--guides", type=int, default=6)
    p_seed.add_argument("--tourists", type=int, default=18)
    p_seed.set_defaults(func=cmd_seed)

    p_storm = sub.add_parser("storm", help="Fire a generic mixed-traffic concurrent burst")
    p_storm.add_argument("--requests", type=int, default=3000)
    p_storm.add_argument("--concurrency", type=int, default=60)
    p_storm.set_defaults(func=cmd_storm)

    p_checkout = sub.add_parser("checkout-stress", help="Hammer checkout specifically (payments+RabbitMQ+tours)")
    p_checkout.add_argument("--requests", type=int, default=500)
    p_checkout.add_argument("--concurrency", type=int, default=40)
    p_checkout.set_defaults(func=cmd_checkout_stress)

    p_throttle = sub.add_parser("throttle-demo", help="Live-demo the progressive login backoff")
    p_throttle.add_argument("--attempts", type=int, default=6)
    p_throttle.set_defaults(func=cmd_throttle_demo)

    p_all = sub.add_parser("all", help="seed, then storm - one command for a live demo")
    p_all.add_argument("--guides", type=int, default=6)
    p_all.add_argument("--tourists", type=int, default=18)
    p_all.add_argument("--requests", type=int, default=3000)
    p_all.add_argument("--concurrency", type=int, default=60)

    def cmd_all(args):
        cmd_seed(args)
        cmd_storm(args)
    p_all.set_defaults(func=cmd_all)

    args = parser.parse_args()
    if args.base_url:
        BASE_URL = args.base_url
    print(f"Target: {BASE_URL}\n")
    args.func(args)


if __name__ == "__main__":
    main()
