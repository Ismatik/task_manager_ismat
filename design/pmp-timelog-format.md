# PMP KIT timelog — deterministic format

The generator counterpart to [`SKILL.md`](./SKILL.md). `SKILL.md` is an **interactive**
skill: it asks the user what was done and for how long. This file specifies a
**generator**: every input comes from the database, from `nodes` and their tracked
`time_entries`, and nothing is asked and nothing is guessed.

Decided in [`PLAN.md` §7, D4](../PLAN.md). **Stage 6 implements it; this file is the
specification only.** Prose is in English because it is read by whoever builds it; every
literal that reaches the platform — field names, activities, the output block — is in
Russian and is reproduced verbatim, because it is copied into a Russian-language form.

---

## 1. The target

The platform is PMP KIT (OpenProject, `pmp.kit.tj`), form **«Добавить трудозатраты»**.
It has exactly four fields:

| Field | Filled from |
|---|---|
| **Дата** | the date of the `time_entries` being logged, `ГГГГ-ММ-ДД` |
| **Часы** | the measured total, rounded (§4) and rendered in platform format |
| **Деятельность** | the node's `activity` (§3) |
| **Комментарий** | generated from the node, then edited by the user (§5) |

### Output block

One record is emitted in exactly this shape (`SKILL.md` §3):

```
Дата: 2026-07-22
Часы: 2h
Деятельность: Разработка

Комментарий:
<краткое описание>

<детальное описание — если запрошено>
```

If only the short description was requested, the detailed block is **not emitted at
all** — not emitted empty, not emitted as a blank line.

---

## 2. Inputs

| Source | Used for |
|---|---|
| `time_entries` (`started_at`, `ended_at`, `node_id`) | Дата, Часы |
| `nodes.activity` | Деятельность |
| `nodes.type` | the default for `activity` when it is unset (§3) |
| `nodes.title` | the short description |
| `nodes.description_md` | the detailed description |

There is no other input. In particular there is **no user prompt** anywhere in the
generation path — see §8.

---

## 3. `activity`

A node carries the field **`activity TEXT`**. Its value is restricted to exactly these
seven, spelled exactly like this, and nothing else is a legal value:

1. `Разработка`
2. `Анализ`
3. `Тестирование`
4. `Документация`
5. `Совещание`
6. `Согласование`
7. `Управление проектом`

### Defaults by node type

When `activity` has not been set on a node, it defaults from `type`. The default is a
starting value, **not a constraint**: the user may change `activity` to any of the seven
at any time, and the override sticks.

| `type` | default `activity` |
|---|---|
| `task` | `Разработка` |
| `bug` | `Тестирование` |
| `project` | `Управление проектом` |
| `note` | `Документация` |

**`habit` is out of scope for time logging.** Habits are checked off against scheduled
occurrences, not timed, so they have no `activity` default and never produce a timelog
record. A habit that somehow carries `time_entries` is a bug, not an input.

---

## 4. Hours

### Hours come only from measured `time_entries`

Never invented, never estimated, never inferred from "this looks like about a day's
work". This is `SKILL.md`'s **«Никогда не выдумывать»** rule, and in the generator it is
structural rather than a matter of discipline: the only number available is the sum of
measured intervals. A node with no `time_entries` in the period produces **no record**,
not a record with a plausible number in it.

The generator also never proposes more hours than were measured.

### Rounding — to the nearest 15 minutes

Rounding is applied **once, to the total** of the entries being aggregated into a single
record — never to each entry before summing. Otherwise six genuine five-minute entries
would each round to zero and half an hour of measured work would disappear.

Ties round **up**: the boundary sits at 7½ minutes past a quarter, so a remainder of
**`:07` rounds down** and **`:08` rounds up**.

### Rendering

Platform format, as in `SKILL.md`: `2h`, `1.5h`, `30m`.

- under 60 minutes → whole minutes, `Nm`
- 60 minutes or more → hours, decimal, trailing zeros trimmed

Because the value is already quantised to 15 minutes, the fractional part is always
`.25`, `.5`, `.75` or absent.

### Worked example

| Measured | Rounded | Rendered | Why |
|---|---|---|---|
| 4 min | 0 | *(no record)* | rounds to zero — nothing to log |
| 7 min | 0 | *(no record)* | **boundary:** remainder 7 ≤ 7½, rounds down |
| 8 min | 15 min | `15m` | **boundary:** remainder 8 > 7½, rounds up |
| 22 min | 15 min | `15m` | 7 past the quarter — down |
| 23 min | 30 min | `30m` | 8 past the quarter — up |
| 30 min | 30 min | `30m` | exact |
| 38 min | 45 min | `45m` | 8 past — up |
| 53 min | 60 min | `1h` | 8 past — up, and now rendered in hours |
| 67 min | 60 min | `1h` | 7 past — down |
| 88 min | 90 min | `1.5h` | |
| 97 min | 90 min | `1.5h` | 7 past — down |
| 98 min | 105 min | `1.75h` | 8 past — up |
| 125 min | 120 min | `2h` | |
| 128 min | 135 min | `2.25h` | 8 past — up |

An entry whose measured total rounds to zero is dropped, and its zero does not
contribute to the total in §6.

---

## 5. Комментарий is generated, then user-edited

The generator produces the text; **the user may rewrite any of it before copying it into
the form**. Generated is the starting point, not the last word. Nothing is sent
anywhere — the user copies the block by hand — so an edit costs nothing and is expected.

### Short description — from `title`

- one line, **≤ 80 characters**, **no trailing period**
- derived from the node's `title`; if the title is longer than 80 characters it is cut
  at a word boundary rather than mid-word
- if the node belongs to a numbered work package (e.g. `#1766`), the number goes at the
  **start** of the short description (`SKILL.md` «Замечания»)

### Detailed description — from `description_md`

- **2–5 sentences**: what was done and what result it produced
- derived from the node's `description_md`, with Markdown syntax stripped — the platform
  field is plain text
- obvious steps are not enumerated
- omitted entirely when only the short description was requested, or when
  `description_md` is empty

### Text rules, inherited from `SKILL.md` §2

- **Impersonal and result-oriented**: «Реализована классификация ошибок», not «я сделал
  классификацию».
- **No filler**: never «проведена работа по», never «осуществлено выполнение
  мероприятий».
- **Technical terms are not translated**: middleware, pull request, миграция, API,
  деплой.

---

## 6. Several records

When the period covers more than one day, more than one node, or more than one
`activity`, each combination is emitted as its own block, in the output format of §1, in
chronological order. After the last block comes **итого часов**.

```
Дата: 2026-07-22
Часы: 2h
Деятельность: Разработка

Комментарий:
Классификация ошибок импорта по кодам платформы

---

Дата: 2026-07-23
Часы: 1.5h
Деятельность: Тестирование

Комментарий:
Регрессия импорта на граничных наборах данных

---

Итого часов: 3.5h
```

The total is the sum of the **already-rounded** per-record values, so that it always
equals what was actually written into the records above it. Summing the raw minutes and
rounding at the end would produce a total that does not match its own parts.

---

## 7. Grouping

One record per `(Дата, node, Деятельность)`. Entries that share all three are summed and
rounded together (§4). Entries on the same day for different nodes stay separate, even
when the activity matches, because their Комментарий comes from different nodes.

---

## 8. What differs from `SKILL.md`

`SKILL.md` **Шаг 1** is the clarifying-questions step: period, hours, language, level of
detail, activity. **None of it applies here.** The generator asks nothing:

| `SKILL.md` asks | The generator reads |
|---|---|
| Период | the period the user already selected in the UI |
| Часы | the measured `time_entries` — the one thing it may never ask for or invent |
| Вид деятельности | `nodes.activity`, defaulted from `nodes.type` (§3) |
| Объём описания | a UI toggle: short only, or short + detailed |
| Язык отчёта | the app's `language` setting |

`SKILL.md`'s note about not having access to past chats is likewise moot: the database
*is* the record of what was done, and it is complete and measured.

What is **kept** from `SKILL.md`: the four form fields, the output block of §3, the seven
activities, the text rules, the `#1766` work-package convention, **итого часов** for
multi-record output, and «Никогда не выдумывать» — which here means the hours are
whatever was measured and nothing else.
