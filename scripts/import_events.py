# Work in progress

import csv

r = open('/home/samer/Downloads/Transposed Events Data 3-22-2017 - Sheet1 (1).csv', 'rb')

reader = csv.reader(r, delimiter=',', quotechar='"')

data = []
last_event = None

reader.next()

for row in reader:
    name, date, event_title, event_type  = row[0], row[1], row[2], row[3]
    if name == '' or date == '' or event_title == '' or event_type == '':
        break

    if (last_event is None or
        last_event['title'] != event_title or last_event['date'] != date or last_event['type'] != event_type):
        if last_event is not None:
            print '|{}|{}|'.format(last_event['type'], event_type)
        last_event = {'title': event_title, 'date': date, 'type': event_type, 'attendees': [name]}
        data.append(last_event)
    else:
        last_event['attendees'].append(name)

print data

