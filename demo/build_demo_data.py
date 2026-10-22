"""Builds demo/transactions.csv: 4 months x 50 synthetic transactions for the demo.

Usage: python3 demo/build_demo_data.py <transactions-synthetic.csv> demo/transactions.csv

The source CSV has only a description and a category per row (balanced: 4000 rows per category), so
dates and amounts are inferred here. The sample is NOT proportional to the source, because that is
uniform across categories. Instead each month has a realistic mix: 9 recurring entries that repeat
every month (rent, insurance, 3 utilities, 4 subscriptions) plus 41 everyday ones.
Only debits are used (the pipeline drops money in). Deterministic: the same inputs give the same file.
"""
import csv
import random
import re
import sys
from collections import defaultdict

SEED = 20261021
MONTHS = [(2026, 6), (2026, 7), (2026, 8), (2026, 9)]
RECURRING = {"Rent": 1, "Insurance": 1, "Utilities": 3, "Subscription": 4}
VARIABLE = {"Restaurants": 10, "Groceries": 7, "Shopping": 6, "Transportation": 4, "Entertainment": 3,
            "Healthcare": 2, "Personal Care": 2, "Education": 1, "Travel": 1, "Fees": 2, "Transfer": 3}
assert sum(RECURRING.values()) + sum(VARIABLE.values()) == 50

# (median, low, high) in dollars for log-normal-ish amounts
AMOUNT_RANGE = {"Restaurants": (18, 6, 95), "Groceries": (55, 12, 190), "Shopping": (38, 9, 260),
                "Transportation": (28, 4, 90), "Entertainment": (22, 6, 120), "Healthcare": (45, 15, 320),
                "Personal Care": (28, 8, 110), "Education": (60, 20, 420), "Travel": (140, 40, 650)}
SUBSCRIPTION_PRICES = [4.99, 7.99, 9.99, 10.99, 12.99, 15.49, 17.99, 22.99]
FEE_AMOUNTS = [3.00, 5.00, 8.00, 12.00, 15.00, 25.00, 35.00]
TRANSFER_AMOUNTS = [10, 15, 20, 25, 40, 50, 60, 75, 100, 150, 200]
UTILITY_DAYS = [12, 15, 19]
SUBSCRIPTION_DAYS = [3, 9, 17, 24]


def load_debits(path):
    by_category = defaultdict(dict)  # category -> normalized description -> description
    with open(path, newline="", encoding="utf-8-sig") as f:
        for row in csv.DictReader(f):
            if row["description"].startswith("[debit] "):
                title = row["description"][len("[debit] "):].strip()
                by_category[row["category"]].setdefault(re.sub(r"\s+", " ", title.lower()), title)
    return {category: sorted(titles.values()) for category, titles in by_category.items()}


def amount_in_range(rng, category):
    median, low, high = AMOUNT_RANGE[category]
    value = rng.lognormvariate(0, 0.6) * median
    return round(min(max(value, low), high), 2)


def build(source_path):
    rng = random.Random(SEED)
    debits = load_debits(source_path)
    rows = []  # (year, month, day, title, amount, source_category)

    recurring = {}
    for category, count in RECURRING.items():
        recurring[category] = rng.sample(debits[category], count)
    rent_amount = float(rng.randrange(1450, 2300, 25))
    insurance_amount = float(rng.randrange(85, 220, 5))
    utility_base = [rng.uniform(45, 140) for _ in range(3)]
    subscription_price = rng.sample(SUBSCRIPTION_PRICES, 4)

    variable = {category: rng.sample(debits[category], count * len(MONTHS)) for category, count in VARIABLE.items()}

    for index, (year, month) in enumerate(MONTHS):
        rows.append((year, month, 1, recurring["Rent"][0], rent_amount, "Rent"))
        rows.append((year, month, 8, recurring["Insurance"][0], insurance_amount, "Insurance"))
        for title, day, base in zip(recurring["Utilities"], UTILITY_DAYS, utility_base):
            rows.append((year, month, day, title, round(base * rng.uniform(0.9, 1.1), 2), "Utilities"))
        for title, day, price in zip(recurring["Subscription"], SUBSCRIPTION_DAYS, subscription_price):
            rows.append((year, month, day, title, price, "Subscription"))
        for category, count in VARIABLE.items():
            for title in variable[category][index * count:(index + 1) * count]:
                if category == "Fees":
                    amount = rng.choice(FEE_AMOUNTS)
                elif category == "Transfer":
                    amount = float(rng.choice(TRANSFER_AMOUNTS))
                else:
                    amount = amount_in_range(rng, category)
                rows.append((year, month, rng.randint(1, 28), title, amount, category))
    rows.sort(key=lambda r: (r[0], r[1], r[2]))  # stable: same-day rows keep their order
    return rows


def main():
    source, target = sys.argv[1], sys.argv[2]
    rows = build(source)
    with open(target, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["date", "amount", "title", "source_category"])
        for year, month, day, title, amount, category in rows:
            writer.writerow([f"{month:02d}/{day:02d}/{year % 100:02d}", f"{-amount:.2f}", title, category])
    print(f"wrote {len(rows)} rows to {target}")


if __name__ == "__main__":
    main()
