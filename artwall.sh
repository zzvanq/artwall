#!/bin/bash

wallpapers=$HOME/Downloads/wallpapers/
mkdir -p wallpapers

used_name=".used"
used_path=$wallpapers$used_name
touch $used_path

current=$(gsettings get org.gnome.desktop.background picture-uri)
current=${current%\'}
current=${current#\'file://}

get_all_wallpapers() {
  find $wallpapers -type f -not -name "$used_name" | sort
}
readarray -t files < <(comm -23 \
  <(get_all_wallpapers) \
  <(cat "$used_path" | sort))
if [ ${#files[@]} -eq 0 ]; then
  # TODO: run artwall-dl to download new files
  > $used_path
  readarray -t files < <(get_all_wallpapers)
  next_idx=0
else
  for i in "${!files[@]}"; do
    if [ "${files[$i]}" = "$current" ]; then
      next_idx=$(( $i + 1 ))
      break
    fi
  done
fi 

next_file=${files[$next_idx]}
if [ ! -f "$next_file" ]; then
  echo "File '$next_file' not found."
  exit 1
fi

echo "Next wallpaper is '$next_file'"
gsettings set org.gnome.desktop.background picture-uri "file://$next_file"
gsettings set org.gnome.desktop.background picture-uri-dark "file://$next_file"
echo $next_file >> $used_path

