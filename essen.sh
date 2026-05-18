curl -s https://www.stwhh.de/gastronomie/mensen-cafes-weiteres/mensa/mensa-harburg | awk '/h5/{getline; gsub(/^[ \t]+|<\/div>/, ""); printf "%s \n", $0} END {print ""}'
