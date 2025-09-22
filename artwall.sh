#!/bin/bash

BLUR=5
TOP_SKIP=0
WALLPAPERS="${XDG_DATA_HOME:-$HOME/.local/share}/artwall/"
mkdir -p $WALLPAPERS

USED_NAME=".used"
USED_PATH=$WALLPAPERS$USED_NAME
touch $USED_PATH

get_all_wallpapers() {
  find $WALLPAPERS -type f -not -name "$USED_NAME" | sort
}
readarray -t files < <(comm -23 \
  <(get_all_wallpapers) \
  <(cat "$USED_PATH" | sort))
if [ ${#files[@]} -eq 0 ]; then
  # TODO: run artwall-dl to download new files
  > $USED_PATH
  readarray -t files < <(get_all_wallpapers)
fi 

# TODO: this check will not make sense after artwall-dl addition
next_file=${files[0]}
if [ ! -f "$next_file" ]; then
  echo "File '$next_file' not found."
  exit 1
fi

echo "Next wallpaper is '$(basename "$next_file")'"
screen_res=$(xrandr | grep '*' | head -1 | awk '{print $1}' | sed 's/\(.*\)x\(.*\)/\1:\2/')
screen_w=${screen_res%:*}
screen_h=${screen_res#*:}
picture_res="$screen_w:$(($screen_h - $TOP_SKIP))"
# TODO: Read metadata from file attrs, if exists - apply
ffmpeg -i "$next_file" -loglevel error \
  -filter_complex "[0:v]scale=$screen_res,gblur=sigma=$BLUR:steps=2[bg];
                   [0:v]scale=$picture_res:force_original_aspect_ratio=decrease[fg];
                   [bg][fg]overlay=(W-w)/2:H-h" \
  -c:v png -f image2 -update 1 -y "$HOME/.config/background" 
echo $next_file >> "$USED_PATH"

