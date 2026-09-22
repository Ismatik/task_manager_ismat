#!/usr/bin/env bash
#
# Nexus — photograph the running binary in every palette, theme and language.
#
# Run it through `make shots`, which is where the WHY of every environment
# variable below is written down. This file is the HOW, and it deliberately does
# not restate the Makefile's explanation.
#
# What it is not: a gate, and not a test (TASKS.md Dev rule 20). The only things
# it decides automatically are "a window existed", "the frame is not blank" and
# "the frame is the right size". Whether anything clips is still a judgement
# made by a person looking at the output.

set -euo pipefail

SIZE="${SHOTS_SIZE:-1024x768}"
OUT="${SHOTS_DIR:-build/shots}"
BINARY="${SHOTS_BINARY:-build/bin/nexus}"
# Seconds to wait for a window to appear, and to let it paint once it has.
WINDOW_TIMEOUT="${SHOTS_WINDOW_TIMEOUT:-25}"
SETTLE="${SHOTS_SETTLE:-4}"

# Every PID this script started, and nothing else. A stale nexus holding the
# single-instance socket makes every later launch exit 0 through the handoff and
# the next capture silently photographs a bare root window — so cleanup matters
# here — but killing a process we did not start would be worse than a bad
# screenshot.
XVFB_PID=""
APP_PID=""

cleanup() {
  stop_app
  if [ -n "$XVFB_PID" ]; then
    kill "$XVFB_PID" 2>/dev/null || true
    wait "$XVFB_PID" 2>/dev/null || true
    XVFB_PID=""
  fi
}
trap cleanup EXIT INT TERM

die() {
  echo >&2
  echo "  make shots FAILED — $*" >&2
  exit 1
}

require() {
  command -v "$1" >/dev/null 2>&1 ||
    die "$1 is not installed. make shots needs Xvfb, ImageMagick (import, convert) and xwininfo."
}

# --------------------------------------------------------------------------
# The two mandatory checks (D31, E4). Neither alone distinguishes the two
# failures: %k cannot say WHY it is 1, and a window that exists can still paint
# blank.

# Reports the number of distinct colours in a PNG. 1 means nothing painted.
colours_in() {
  convert "$1" -format "%k" info:
}

dimensions_of() {
  convert "$1" -format "%wx%h" info:
}

windows_on_display() {
  xwininfo -display "$DISPLAY" -root -children 2>/dev/null | grep -c -i nexus || true
}

# --------------------------------------------------------------------------

stop_app() {
  if [ -n "$APP_PID" ]; then
    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true
    APP_PID=""
  fi
}

# Picks a display number no X server is answering on. Never reuses one that is
# already up: that would be somebody else's server, and the trap above would
# then be asked to kill something this script did not start.
free_display() {
  local n
  for n in $(seq 99 120); do
    if ! xdpyinfo -display ":$n" >/dev/null 2>&1; then
      echo ":$n"
      return 0
    fi
  done
  die "no free X display between :99 and :120"
}

capture_state() {
  local fixture="$1" palette="$2" theme="$3" language="$4"
  local name="$fixture-$palette-$theme-$language"
  local png="$OUT/$name.png"
  local data="$PWD/$OUT/data/$fixture"
  local runtime="$PWD/$OUT/run"

  # Everything this harness claims about never opening the user's database
  # rests on the line above. internal/store/db.go honours XDG_DATA_HOME only
  # when it is absolute and otherwise falls back to ~/.local/share WITHOUT
  # failing, so a relative path here would not break the run: it would succeed
  # against the wrong database, seed rows into it, and walk past both checks
  # below with a window and a non-blank frame. A comment cannot catch that
  # (D17), so the next line is an assertion instead of a note. $runtime is
  # deliberately not covered: platform.RuntimeDir() takes XDG_RUNTIME_DIR as
  # given, so a relative one resolves against this process's directory and
  # lands under the repo, which cannot be anyone else's socket.
  case "$data" in
  /*) ;;
  *) die "XDG_DATA_HOME would not be absolute: $data
      internal/store/db.go ignores a relative XDG_DATA_HOME and falls back to
      ~/.local/share/nexus/nexus.db, so this run would seed and photograph the
      user's own database instead of a throwaway one." ;;
  esac

  mkdir -p "$data" "$runtime"
  # Only ever inside our own throwaway runtime directory.
  rm -f "$runtime/nexus/ipc.sock"

  local board=()
  [ "$fixture" = "board" ] && board=(-board)

  XDG_DATA_HOME="$data" go run -tags shots ./scripts/shots \
    -palette "$palette" -theme "$theme" -language "$language" "${board[@]}" ||
    die "seeding $name failed"

  # THE INVOCATION. See the `shots:` target in the Makefile for why each of
  # these is here; the short version is that GDK_BACKEND=x11 and the absence of
  # WAYLAND_DISPLAY are what put the window on $DISPLAY at all.
  env -u WAYLAND_DISPLAY \
    DISPLAY="$DISPLAY" \
    GDK_BACKEND=x11 \
    LC_NUMERIC=C \
    XDG_DATA_HOME="$data" \
    XDG_RUNTIME_DIR="$runtime" \
    "$BINARY" >"$OUT/$name.log" 2>&1 &
  APP_PID=$!

  # CHECK 1 — window presence, first, because it is cheap and specific.
  local waited=0 windows=0
  while [ "$waited" -lt "$WINDOW_TIMEOUT" ]; do
    windows=$(windows_on_display)
    [ "$windows" -gt 0 ] && break
    sleep 1
    waited=$((waited + 1))
  done

  if [ "$windows" -eq 0 ]; then
    stop_app
    die "no window was ever created on $DISPLAY for the state '$name'.
      This is NOT a blank frame — it is a window that went somewhere else, and
      it has a different fix. Check GDK_BACKEND and WAYLAND_DISPLAY: with
      WAYLAND_DISPLAY set, GTK prefers the Wayland backend and the window opens
      on the real compositor, leaving $DISPLAY empty while the process runs
      healthily. This target sets GDK_BACKEND=x11 and passes 'env -u
      WAYLAND_DISPLAY'; if you have removed either, that is why.
      The process log is $OUT/$name.log."
  fi

  sleep "$SETTLE"
  import -display "$DISPLAY" -window root "$png"
  stop_app

  # CHECK 2 — blankness, second, because a present window can still paint blank.
  local k
  k=$(colours_in "$png")
  if [ "$k" -le 1 ]; then
    die "$png is BLANK — $k distinct colour — for the state '$name'.
      A window existed on $DISPLAY, so this is a rendering failure and not a
      backend one. The process log is $OUT/$name.log."
  fi

  local got
  got=$(dimensions_of "$png")
  [ "$got" = "$SIZE" ] ||
    die "$png is $got, not the $SIZE this matrix is about."

  printf '  %-34s %s  %s colours\n' "$name" "$got" "$k"
}

# --------------------------------------------------------------------------

for tool in Xvfb import convert xwininfo xdpyinfo; do require "$tool"; done
[ -x "$BINARY" ] || die "$BINARY is missing. Run 'make build' first."

rm -rf "$OUT"
mkdir -p "$OUT"

# PROVE THE BLANKNESS FLOOR IS LIVE, before anything is captured.
#
# Check 2 lives behind check 1: if a window is always found, the %k floor is
# never the thing that fires, and a floor nobody ever reaches is indistinguishable
# from a floor that does not work (D32). So it is exercised here on a PNG that is
# blank by construction — every run, not once in a commit message.
BLANK="$OUT/.blankness-floor.png"
convert -size "$SIZE" xc:black "$BLANK"
[ "$(colours_in "$BLANK")" -le 1 ] ||
  die "the blankness floor is broken: a solid black $SIZE PNG did not read as 1 colour.
      Every 'not blank' result below would be meaningless."
rm -f "$BLANK"

DISPLAY=$(free_display)
export DISPLAY
Xvfb "$DISPLAY" -screen 0 "${SIZE}x24" -nolisten tcp >"$OUT/xvfb.log" 2>&1 &
XVFB_PID=$!

waited=0
until xdpyinfo -display "$DISPLAY" >/dev/null 2>&1; do
  sleep 1
  waited=$((waited + 1))
  [ "$waited" -lt 15 ] || die "Xvfb did not come up on $DISPLAY (see $OUT/xvfb.log)"
done

echo "Capturing at $SIZE on $DISPLAY, into $OUT/"
echo

# The palette/theme/language sets come from internal/domain, printed by the seed
# program. The two FIXTURES are this script's own idea and are named here: an
# empty board is the state K15 is about, and a board with a card in every column
# is the state K8 was about.
while read -r palette theme language; do
  [ -n "$palette" ] || continue
  for fixture in empty board; do
    capture_state "$fixture" "$palette" "$theme" "$language"
  done
done < <(go run -tags shots ./scripts/shots -matrix)

cleanup
rm -rf "$OUT/data" "$OUT/run"

echo
echo "$(ls "$OUT"/*.png | wc -l) frames in $OUT/ — every one non-blank and $SIZE."
echo "A capture is not a test: 'nothing clips' is still a judgement made by looking."
