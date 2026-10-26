#!/usr/bin/env python3
"""Generates four synthetic Bank of America style statements from demo/transactions.csv.

Run from the repo root (needs reportlab: python3 -m venv /tmp/v && /tmp/v/bin/pip install reportlab):
    /tmp/v/bin/python demo/statements/build_statements.py
Writes demo/statements/statement-2026-{06,07,08,09}.pdf. Output is deterministic
(reportlab invariant mode). Fonts: macOS system fonts (see below); the PDFs are
committed so the demo does not depend on regenerating them. After any change re-run
`go test -run DemoStatements .` to confirm the real OCR pipeline still reads every row exactly. All names and numbers are fictional.
"""
import csv
import os
import re
import tempfile
from decimal import Decimal

from reportlab.lib.colors import HexColor, black
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas

HERE = os.path.dirname(os.path.abspath(__file__))
CSV = os.path.join(HERE, "..", "transactions.csv")
RED = HexColor("#E31837")
GREY = HexColor("#555555")
NAME = "JAMIE EXAMPLE"
ACCOUNT = "1234 5678 9012"
# Tesseract confuses some glyphs (0/O, 1/l), so the font choice matters. macOS system fonts are used
# (Verdana for text; the narrow Arial zero so 0 is never read as O). Without them it falls back to
# Helvetica, which still renders but is NOT guaranteed to OCR exactly.
_SYS = "/System/Library/Fonts/Supplemental/"
FONT, BOLD, ZERO, ONE = "Helvetica", "Helvetica-Bold", "Helvetica", "Helvetica"
if os.path.exists(_SYS + "Verdana.ttf"):
    pdfmetrics.registerFont(TTFont("Verdana", _SYS + "Verdana.ttf"))
    pdfmetrics.registerFont(TTFont("Verdana-Bold", _SYS + "Verdana Bold.ttf"))
    pdfmetrics.registerFont(TTFont("ArialNarrow", _SYS + "Arial Narrow.ttf"))
    pdfmetrics.registerFont(TTFont("Trebuchet", _SYS + "Trebuchet MS.ttf"))
    FONT, BOLD, ZERO = "Verdana", "Verdana-Bold", "ArialNarrow"
    ONE = "Trebuchet"
SIZE = 10.5
LEFT, DESC_X, RIGHT = 54, 112, 558
DESC_MAX = 300  # points available for the first description line (before the amount)
LINE = 17
TOP_Y = 792 - 112  # first red heading baseline
MAX_BODY_LINES = 29  # keeps each statement at 2-3 transaction pages
MONTHS = {6: "June", 7: "July", 8: "August", 9: "September"}
LAST_DAY = {6: 30, 7: 31, 8: 31, 9: 30}

OTHER = re.compile(
    r"PPD ID:|WEB ID:|Zelle payment to|venmo payment to|CASH APP PAYMENT TO|Online Banking transfer|"
    r"PAYPAL INST XFER|DDA TO DDA|Manual Db|BUYER PAYMENT:|withdrawal", re.I)

DEPOSITS = {
    6: [("06/05/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("06/19/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("06/24/26", "REFUND STORE 4411 PURCHASE RETURN", "37.82")],
    7: [("07/03/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("07/17/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("07/28/26", "Online Banking transfer from SAV 4321 Confirmation# 5550123", "200.00")],
    8: [("08/07/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("08/21/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00")],
    9: [("09/04/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("09/18/26", "ACME CORP PAYROLL PPD ID: 1234567890", "2450.00"),
        ("09/25/26", "REFUND ONLINE ORDER 7781 RETURN CREDIT", "64.15")],
}


def money(d, sign=""):
    return f"{sign}{abs(d):,.2f}"


def load(month):
    rows = []
    with open(CSV, newline="") as f:
        for r in csv.DictReader(f):
            if int(r["date"][:2]) == month and r["date"].endswith("26"):
                rows.append((r["date"], Decimal(r["amount"]), r["title"]))
    rows.sort(key=lambda r: r[0])  # stable: keeps CSV order within a day
    return rows


def wrap(title, width):
    """Split at spaces: first line fits `width`, the rest wraps on continuation lines."""
    words, lines, cur = title.split(" "), [], ""
    limit = width
    for w in words:
        t = w if not cur else cur + " " + w
        if cur and pdfmetrics.stringWidth(t, FONT, SIZE) > limit:
            lines.append(cur)
            cur, limit = w, RIGHT - DESC_X
        else:
            cur = t
    lines.append(cur)
    return lines


class Doc:
    def __init__(self, path, month):
        self.c = canvas.Canvas(path, pagesize=(612, 792), invariant=1)
        self.month, self.page, self.y, self.used = month, 1, 0, 0
        self.total_pages = 0

    def text(self, x, y, s, font=FONT, size=SIZE, color=black, right=False, desc=False):
        runs = []  # (font, size, text); "0" and "1" in descriptions use fonts OCR reads unambiguously
        for ch in s:
            f, z = font, size
            if not desc:
                pass
            elif font == FONT and ch == "0":
                f, z = ZERO, size
            elif font == FONT and ch == "1":
                f = ONE
            if runs and runs[-1][0] == f and runs[-1][1] == z:
                runs[-1][2] += ch
            else:
                runs.append([f, z, ch])
        if right:
            x -= sum(pdfmetrics.stringWidth(t, f, z) for f, z, t in runs)
        self.c.setFillColor(color)
        for f, z, t in runs:
            self.c.setFont(f, z)
            self.c.drawString(x, y, t)
            x += pdfmetrics.stringWidth(t, f, z)

    def footer(self):
        self.text(306, 40, f"Page {self.page} of {self.total_pages}", size=9, color=GREY)

    def header(self):
        m = MONTHS[self.month]
        self.text(LEFT, 792 - 50, f"{NAME}  |  Account # {ACCOUNT}  |  for {m} 1, 2026 to {m} {LAST_DAY[self.month]}, 2026",
                  size=9, color=GREY)
        c = self.c  # red logo flag, right side only (x > 60 pt)
        c.setFillColor(RED)
        c.rect(520, 792 - 62, 38, 14, fill=1, stroke=0)
        self.y, self.used = TOP_Y, 0

    def new_page(self):
        self.footer()
        self.c.showPage()
        self.page += 1
        self.header()

    def heading(self, s, size=12):
        self.text(LEFT, self.y, s, BOLD, size, RED)
        self.y -= LINE + 3
        self.used += 1

    def table_header(self):
        self.text(LEFT, self.y, "Date Description Amount", FONT, SIZE)
        self.y -= LINE
        self.used += 1

    def room(self, n):
        return self.used + n <= MAX_BODY_LINES

    def section(self, sub, rows, total_label, top_first=None, top_cont=None):
        """rows: (date, signed Decimal, text). Breaks pages with continuation headings."""
        first = True
        pending = list(rows)
        while True:
            if first:
                if top_first:
                    self.heading(top_first)
                if sub:
                    self.heading(sub, 11)
            else:
                if top_cont:
                    self.heading(top_cont + " - continued")
                if sub:
                    self.heading(sub + " - continued", 11)
            self.table_header()
            first = False
            broke = False
            while pending:
                d, amt, t = pending[0]
                lines = wrap(t, DESC_MAX)
                if not self.room(len(lines) + 2):
                    broke = True
                    break
                pending.pop(0)
                self.text(LEFT, self.y, d)
                self.text(DESC_X, self.y, lines[0], desc=True)
                self.text(RIGHT, self.y, money(amt, "-" if amt < 0 else ""), right=True)
                self.y -= LINE
                for extra in lines[1:]:
                    self.text(DESC_X, self.y, extra, desc=True)
                    self.y -= LINE
                self.used += len(lines)
            if not broke:
                total = sum(r[1] for r in rows)
                self.y -= 4
                self.text(LEFT, self.y, f"{total_label} {'-' if total < 0 else ''}${abs(total):,.2f}", BOLD)
                self.y -= LINE + 10
                self.used += 2
                return
            self.text(LEFT, self.y - 4, "continued on the next page", FONT, 9.5, GREY)
            self.new_page()

    def finish(self, blank_pages=2):
        for _ in range(blank_pages):
            self.new_page()
            self.y = 792 - 300
            self.text(306, self.y, "This page intentionally left blank", size=10, right=False)
        self.footer()
        self.c.showPage()
        self.c.save()


def summary_page(d, month, deposits, spending):
    c, m = d.c, MONTHS[month]
    d.text(LEFT, 792 - 60, "Bank of Examples", BOLD, 18, RED)
    d.text(LEFT, 792 - 100, NAME, BOLD, 12)
    d.text(LEFT, 792 - 116, "123 Sample Street, Anytown, ST 00000")
    d.text(LEFT, 792 - 150, f"Your Advantage Checking  |  Account # {ACCOUNT}")
    d.text(LEFT, 792 - 166, f"for {m} 1, 2026 to {m} {LAST_DAY[month]}, 2026")
    d.text(LEFT, 792 - 220, "Account summary", BOLD, 12)
    dep, spend = sum(x[1] for x in deposits), sum(x[1] for x in spending)
    d.text(LEFT, 792 - 245, "Total deposits and other additions")
    d.text(RIGHT, 792 - 245, f"${dep:,.2f}", right=True)
    d.text(LEFT, 792 - 262, "Total withdrawals and other subtractions")
    d.text(RIGHT, 792 - 262, f"-${abs(spend):,.2f}", right=True)
    d.text(LEFT, 792 - 320, "This is a fictional statement generated for demonstration purposes.", size=9, color=GREY)
    d.footer()
    c.showPage()
    d.page += 1


def legal_page(d):
    d.header()
    lines = ["Important information about your account (fictional sample text).",
             "Please review this statement and report any errors within 60 days.",
             "Electronic transfers: contact us right away if you think a transfer is wrong.",
             "This page contains no transaction detail."]
    y = 792 - 130
    for s in lines:
        d.text(LEFT, y, s, size=10)
        y -= 18
    d.footer()
    d.c.showPage()
    d.page += 1


def build(month, total_pages=0, path=None):
    """Writes one statement and returns (path, number of pages). total_pages fills in "Page n of N"."""
    path = path or os.path.join(HERE, f"statement-2026-{month:02d}.pdf")
    spend = load(month)
    assert len(spend) == 50, (month, len(spend))
    deposits = [(d, Decimal(a), t) for d, t, a in DEPOSITS[month]]
    atm = [r for r in spend if not OTHER.search(r[2])]
    other = [r for r in spend if OTHER.search(r[2])]
    d = Doc(path, month)
    d.total_pages = total_pages
    summary_page(d, month, deposits, spend)
    legal_page(d)
    d.header()
    d.section(None, deposits, "Total deposits and other additions", top_first="Deposits and other additions")
    w = "Withdrawals and other subtractions"
    d.section("ATM and debit card subtractions", atm, "Total ATM and debit card subtractions",
              top_first=w, top_cont=w)
    d.section("Other subtractions", other, "Total other subtractions", top_cont=w)
    d.finish()
    return path, d.page


if __name__ == "__main__":
    for m in MONTHS:
        with tempfile.TemporaryDirectory() as scratch:  # first pass only counts the pages
            _, pages = build(m, path=os.path.join(scratch, "count.pdf"))
        print(build(m, total_pages=pages)[0])
