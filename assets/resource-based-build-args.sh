#! usr/bin/env bash

has_from_file() {
  result=$(echo "$1" | jq 'has("from_file")')
  echo "$result"
}

elevate_from_file_kvps() {
  declare -A kvpmap

  while IFS=" " read -r key value
  do
    kvpmap["$key"]="$value"
  done < <(echo "$1" | jq -r '.from_file | to_entries | .[] | .key+" "+.value')

  for key in "${!kvpmap[@]}"
  do
    kvpmap["$key"]=$(cat "${kvpmap["$key"]}")
  done

  rejsoned='{'

  for key in "${!kvpmap[@]}"
  do
    if [[ -n "$has_skipped_first_line" ]]
    then
      rejsoned+=','
    else
      has_skipped_first_line=true
    fi
    rejsoned+='"'
    rejsoned+="$key"
    rejsoned+='"'
    rejsoned+=':'
    rejsoned+='"'
    rejsoned+="${kvpmap[$key]}"
    rejsoned+='"'
  done

  rejsoned+='}'

  no_more_from_file=$(echo "$1" | jq 'del(.from_file)')

  result=$(echo "$no_more_from_file" "$rejsoned" | jq -c -s add)

  echo "$result"
}

