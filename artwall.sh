#!/bin/bash

set -euo pipefail

BLUR=15
TOP_SKIP=31
WALLPAPERS="${XDG_DATA_HOME:-$HOME/.local/share}/artwall/"
mkdir -p $WALLPAPERS
MANAGED="$WALLPAPERS/managed/"
mkdir -p $MANAGED

USED_NAME=".used"
USED_PATH=$WALLPAPERS$USED_NAME
touch $USED_PATH

DOWNLOADER_URL="https://www.nga.gov/search/json?begin_year=522&end_year=1934&f[]=awtype:107231&f[]=movement:82506&f[]=movement:29206&f[]=movement:83431&f[]=movement:33831&f[]=movement:46221&f[]=movement:58341&f[]=movement:3811&f[]=movement:53841&f[]=movement:91366&f[]=movement:76071&page="
DOWNLOAD_AMOUNT=12
CURRENT_NAME=".current"
CURRENT_PATH=$WALLPAPERS$CURRENT_NAME

get_all_wallpapers() {
  find $WALLPAPERS -type f -not -name "$CURRENT_NAME" -not -name "$USED_NAME" | sort
}

readarray -t files < <(comm -23 \
  <(get_all_wallpapers) \
  <(sort "$USED_PATH"))

if [ ${#files[@]} -eq 0 ]; then
  page=$(cat $CURRENT_PATH 2>/dev/null || echo 1)
  artwall-dl -d $MANAGED -n $DOWNLOAD_AMOUNT -l $DOWNLOADER_URL$page
  echo $page > $CURRENT_PATH
  > $USED_PATH
  readarray -t files < <(get_all_wallpapers)
fi 

next_file=${files[0]}
if [ ! -f "$next_file" ]; then
  echo "File '$next_file' not found."
  exit 1
fi

echo "Next wallpaper is '$(basename "$next_file")'"

screen_res=$(xrandr | awk '/\*/{print $1;exit}' | sed 's/x/:/')
screen_w=${screen_res%:*}
screen_h=${screen_res#*:}
picture_res="$screen_w:$(($screen_h - $TOP_SKIP))"

meta=""
author=""
title=""
year=""
description=""
if [ "$(file -b --mime-type "$next_file")" == "image/jpeg" ]; then
  meta=$(tinymedia -i "$next_file" -m "author,title,year,description" -mv "tinymetagzip"  )
  author=$(sed -n 's/"author"="\([^"]*\)"/\1/p' <<< "$meta")
  title=$(sed -n 's/"title"="\([^"]*\)"/\1/p' <<< "$meta")
  year=$(sed -n 's/"year"="\([^"]*\)"/\1/p' <<< "$meta")
  description=$(sed -n 's/"description"="\([^"]*\)"/\1/p' <<< "$meta" | sed 's/\\n/\n\n/g' | sed "s/:/꞉/g" | sed "s/'/’/g"|  fmt -w 70 )
fi

vf=""
box_w=0
if [[ -n "$description" ]]; then
  box_w=710
  box_m="$screen_w-$box_w+$(($box_w/2))-tw/2"
  title_offset=150
  vf="drawtext=text='$title':x=$box_m:y=$TOP_SKIP+$title_offset:fontsize=34:fontcolor=white:borderw=1:bordercolor=black,\
      drawtext=text='$author. $year':x=$box_m:y=$TOP_SKIP+$title_offset+40:fontsize=24:fontcolor=gray:borderw=1:bordercolor=black,\
      drawtext=text='$description':x=$box_m:y=$TOP_SKIP+(H-$TOP_SKIP-th)/2:fontsize=20:fontcolor=white:line_spacing=3:text_align=C+C:borderw=2:bordercolor=black"
fi

ffmpeg -i "$next_file" -loglevel error \
  -filter_complex "[0:v]scale=$screen_res,gblur=sigma=$BLUR:steps=2[bg];
                   [0:v]scale=$picture_res:force_original_aspect_ratio=decrease[fg];
                   [bg][fg]overlay=(W-$box_w-w)/2:$TOP_SKIP,${vf}" \
  -c:v png -f image2 -update 1 -y "$HOME/.config/background" 

echo $next_file >> "$USED_PATH"

