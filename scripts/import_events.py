#!/usr/bin/env python

# Work in progress

# Install pymysql first:
#  - pip install PyMySQL

# To work off of this data in your dev vm, run the following:
#  - go run create_db.go --no-fake-data
#  - python import_events.py


import csv
from datetime import datetime
import pymysql.cursors

def format_date(raw_date):
    return datetime.strptime(raw_date, '%m/%d/%Y').strftime('%Y-%m-%d')

VALID_EVENT_TYPES = set(['Sanctuary', 'Key Event', 'Outreach', 'Community', 'Protest', 'Working Group'])

def get_events_from_csv():
    # Downloaded from here:
    # https://docs.google.com/spreadsheets/d/1MmGNEamRMFfq99gkclWsFN2PTiX71gr0DcEofwK5avg/edit
    r = open('./transposed-events-data-3-22-2017.csv', 'rb')

    reader = csv.reader(r, delimiter=',', quotechar='"')

    events = []
    last_event = None

    reader.next()

    for row in reader:
        name, date, event_title, event_type  = row[0], row[1], row[2], row[3]
        if name == '' or date == '' or event_title == '' or event_type == '':
            break
        date = format_date(date)
        if event_type not in VALID_EVENT_TYPES:
            raise Exception("not a valid event type: %s" % event_type)

        if (last_event is None or last_event['title'] != event_title or last_event['date'] != date or last_event['type'] != event_type):

            last_event = {'title': event_title,
                          'date': date,
                          'type': event_type,
                          'attendees': [name]}
            events.append(last_event)
        else:
            last_event['attendees'].append(name)

    return events

def get_event_key(event_data):
    return event_data['title'] + '__' + event_data['date'] + '__' + event_data['type']

def insert_data_into_db(events):
    connection = pymysql.connect(host='localhost', user='adb_user',
                                 password='adbpassword', db='adb_db',
                                 charset='utf8mb4',
                                 cursorclass=pymysql.cursors.DictCursor)

    c = connection.cursor()

    user_to_id = {}

    print "inserting user data"
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

    print "inserting event and attendance data"
    # Then, insert all events and attendance.
    for event in events:
        # First, see if event already exists.
        c.execute("SELECT id FROM events WHERE name = %s and date = %s and event_type = %s", (event['title'], event['date'], event['type']))
        event_data = c.fetchone()
        if not event_data:
            # Insert the event b/c it doesn't exist.
            c.execute("INSERT INTO events (name, date, event_type) VALUES (%s, %s, %s)", (event['title'], event['date'], event['type']))
            connection.commit()
            c.execute("SELECT id FROM events WHERE name = %s and date = %s and event_type = %s", (event['title'], event['date'], event['type']))
            event_data = c.fetchone()

        assert event_data

        # Delete all event attendance
        c.execute("DELETE FROM event_attendance WHERE event_id = %s", event_data['id'])
        connection.commit()

        # Insert all event attendance
        for user in event['attendees']:
            c.execute("INSERT INTO event_attendance (activist_id, event_id) VALUES (%s, %s)", (user_to_id[user], event_data['id']))
        connection.commit()

def main():
    events = get_events_from_csv()
    #extra_stuff(events)
    insert_data_into_db(events)

if __name__ == '__main__':
    main()
