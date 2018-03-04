$files = gci

foreach ($file in $files) {

    write-host $file
    $rand = get-random

    rename-item $file $rand.jpg

}