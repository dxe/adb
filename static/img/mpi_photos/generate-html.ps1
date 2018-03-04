$files = gci

foreach ($file in $files) {

    #write-host $file
    #$rand = get-random
    #$rand

    #rename-item $file ('' + $rand + '.jpg')

    write-host ('<div class="item"><img src="/static/img/mpi_photos/' + $file + '" style="width:100%;"></div>')

}