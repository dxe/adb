

CREATE TABLE activists (
  id int(8) NOT NULL,
  name varchar(80) NOT NULL,
  email varchar(80) DEFAULT NULL,
  chapter_id int(3) DEFAULT NULL,
  phone varchar(20) DEFAULT NULL,
  city varchar(40) DEFAULT NULL,
  zipcode varchar(15) DEFAULT NULL,
  country varchar(80) DEFAULT NULL,
  facebook varchar(80) DEFAULT NULL,
  exclude_from_leaderboard tinyint(1) NOT NULL DEFAULT '0',
  core_staff tinyint(1) NOT NULL DEFAULT '0',
  global_team_member tinyint(1) NOT NULL DEFAULT '0',
  liberation_pledge tinyint(1) DEFAULT NULL
);


INSERT INTO activists VALUES
  (1, 'Adam Kol', 'adam@directactioneverywhere.com', 2, '9542635719', 'Berkeley', '', 'United States', '', 0, 0, 1, 1);
INSERT INTO activists VALUES
  (2, 'Robin Houseman', 'testtest@gmail.com', 2, '4398943', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (3, 'Jake Hong', 'test@comcast.net', 2, '7077206366', 'Fairfield', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (4, 'Samer Samer', 'test.test.test@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (5, 'Samer Masterson', 'alexis.l.levitt@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);


CREATE TABLE chapters (
  id int(3) NOT NULL,
  name varchar(30) DEFAULT NULL
);

INSERT INTO chapters VALUES
(1, 'Global');
INSERT INTO chapters VALUES
(2, 'SF Bay');

CREATE TABLE event_attendance (
  activist_id int(8) NOT NULL,
  event_id int(8) NOT NULL,
  timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attendance_order int(8) NOT NULL
);

CREATE TABLE events (
  id int(8) NOT NULL,
  name varchar(60) NOT NULL,
  date date NOT NULL,
  type_id int(2) NOT NULL,
  chapter_id int(3) NOT NULL
);

CREATE TABLE event_type (
  id int(2) NOT NULL,
  name varchar(30) NOT NULL
);

INSERT INTO event_type VALUES
(1, 'Working Group');
INSERT INTO event_type VALUES
(2, 'Community');
INSERT INTO event_type VALUES
(3, 'Protest');
INSERT INTO event_type VALUES
(4, 'Key Event');
INSERT INTO event_type VALUES
(5, 'Outreach');
INSERT INTO event_type VALUES
(6, 'Sanctuary');


CREATE TABLE press (
  id int(8) NOT NULL,
  date date NOT NULL,
  outlet varchar(60) NOT NULL,
  url varchar(200) NOT NULL,
  headline int(200) NOT NULL
);

CREATE TABLE working_group_members (
  id int(4) NOT NULL,
  working_group_id int(4) NOT NULL,
  activist_id int(8) NOT NULL,
  leader tinyint(1) NOT NULL DEFAULT '0'
);

INSERT INTO working_group_members VALUES
(1, 1, 1, 1);
INSERT INTO working_group_members VALUES
(2, 1, 2, 0);
INSERT INTO working_group_members VALUES
(3, 2, 1, 0);


CREATE TABLE working_groups (
  id int(4) NOT NULL,
  name varchar(30) DEFAULT NULL,
  short_name varchar(30) NOT NULL,
  chapter int(3) DEFAULT NULL
);


INSERT INTO working_groups VALUES
(1, 'Outreach', 'outreach', 1);
INSERT INTO working_groups VALUES
(2, 'SF Bay Meetups', 'sfbay-meetups', 2);
INSERT INTO working_groups VALUES
(3, 'SF Bay Connections', 'sfbay-connections', 2);
INSERT INTO working_groups VALUES
(4, 'SF Bay Animal Care', 'sfbay-sanctuary', 2);
INSERT INTO working_groups VALUES
(5, 'Press', 'press', 1);
