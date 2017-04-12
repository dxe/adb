# Work in progress

# Install pymysql first:
#  - pip install PyMySQL

import csv
import pymysql.cursors

# Downloaded from here:
# https://docs.google.com/spreadsheets/d/1MmGNEamRMFfq99gkclWsFN2PTiX71gr0DcEofwK5avg/edit
r = open('/home/samer/Downloads/Transposed Events Data 3-22-2017 - Sheet1 (1).csv', 'rb')

reader = csv.reader(r, delimiter=',', quotechar='"')

events = []
last_event = None

reader.next()

for row in reader:
    name, date, event_title, event_type  = row[0], row[1], row[2], row[3]
    if name == '' or date == '' or event_title == '' or event_type == '':
        break

    if (last_event is None or
        last_event['title'] != event_title or last_event['date'] != date or last_event['type'] != event_type):
        last_event = {'title': event_title, 'date': date, 'type': event_type, 'attendees': [name]}
        events.append(last_event)
    else:
        last_event['attendees'].append(name)

connection = pymysql.connect(host='localhost', user='adb_user',
                             password='adbpassword', db='adb_db',
                             charset='utf8mb4',
                             cursorclass=pymysql.cursors.DictCursor)

c = connection.cursor()

user_to_id = {}

# First, insert all attendees
for event in events:
    for name in event['attendees']:
        # Don't insert someone we've already seen
        if name in user_to_id:
            continue

        # See if this person exists in our database.
        c.execute("SELECT id FROM activists WHERE name = %s", name)
        user_data = c.fetchone()

        if not user_data:
            # The user doesn't exist, so insert them and fetch them
            # again.
            c.execute("INSERT INTO activists (name) VALUES (%s)", name)
            connection.commit()
            c.execute("SELECT id FROM activists WHERE name = %s", name)
            user_data = c.fetchone()

        assert user_data
        user_to_id[name] = user_data['id']

print user_to_id
