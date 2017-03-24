

CREATE TABLE Activists (
  ID int(8) NOT NULL,
  Name varchar(80) NOT NULL,
  Email varchar(80) DEFAULT NULL,
  ChapterID int(3) DEFAULT NULL,
  Phone varchar(20) DEFAULT NULL,
  City varchar(40) DEFAULT NULL,
  Zipcode varchar(15) DEFAULT NULL,
  Country varchar(80) DEFAULT NULL,
  Facebook varchar(80) DEFAULT NULL,
  ExcludeFromLeaderBoard tinyint(1) NOT NULL DEFAULT '0',
  CoreStaff tinyint(1) NOT NULL DEFAULT '0',
  GlobalTeamMember tinyint(1) NOT NULL DEFAULT '0',
  LiberationPledge tinyint(1) DEFAULT NULL
);


INSERT INTO Activists VALUES
  (1, 'Adam Kol', 'adam@directactioneverywhere.com', 2, '9542635719', 'Berkeley', '', 'United States', '', 0, 0, 1, 1);
INSERT INTO Activists VALUES
  (2, 'Robin Houseman', 'testtest@gmail.com', 2, '4398943', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO Activists VALUES
  (3, 'Jake Hong', 'test@comcast.net', 2, '7077206366', 'Fairfield', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO Activists VALUES
  (4, 'Samer Samer', 'test.test.test@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO Activists VALUES
  (5, 'Samer Masterson', 'alexis.l.levitt@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);


CREATE TABLE Chapters (
  ID int(3) NOT NULL,
  Name varchar(30) DEFAULT NULL
);

INSERT INTO Chapters VALUES
(1, 'Global');
INSERT INTO Chapters VALUES
(2, 'SF Bay');

CREATE TABLE EventAttendance (
  ActivistID int(8) NOT NULL,
  EventID int(8) NOT NULL,
  Timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  AttendanceOrder int(8) NOT NULL
);

CREATE TABLE Events (
  ID int(8) NOT NULL,
  Name varchar(60) NOT NULL,
  Date date NOT NULL,
  TypeID int(2) NOT NULL,
  ChapterID int(3) NOT NULL
);

CREATE TABLE EventType (
  ID int(2) NOT NULL,
  Name varchar(30) NOT NULL
);

INSERT INTO EventType VALUES
(1, 'Working Group');
INSERT INTO EventType VALUES
(2, 'Community');
INSERT INTO EventType VALUES
(3, 'Protest');
INSERT INTO EventType VALUES
(4, 'Key Event');
INSERT INTO EventType VALUES
(5, 'Outreach');
INSERT INTO EventType VALUES
(6, 'Sanctuary');


CREATE TABLE Press (
  ID int(8) NOT NULL,
  Date date NOT NULL,
  Outlet varchar(60) NOT NULL,
  URL varchar(200) NOT NULL,
  Headline int(200) NOT NULL
);

CREATE TABLE WorkingGroupMembers (
  ID int(4) NOT NULL,
  WorkingGroupID int(4) NOT NULL,
  ActivistID int(8) NOT NULL,
  Leader tinyint(1) NOT NULL DEFAULT '0'
);

INSERT INTO WorkingGroupMembers VALUES
(1, 1, 1, 1);
INSERT INTO WorkingGroupMembers VALUES
(2, 1, 2, 0);
INSERT INTO WorkingGroupMembers VALUES
(3, 2, 1, 0);


CREATE TABLE WorkingGroups (
  ID int(4) NOT NULL,
  Name varchar(30) DEFAULT NULL,
  ShortName varchar(30) NOT NULL,
  Chapter int(3) DEFAULT NULL
);


INSERT INTO WorkingGroups VALUES
(1, 'Outreach', 'outreach', 1);
INSERT INTO WorkingGroups VALUES
(2, 'SF Bay Meetups', 'sfbay-meetups', 2);
INSERT INTO WorkingGroups VALUES
(3, 'SF Bay Connections', 'sfbay-connections', 2);
INSERT INTO WorkingGroups VALUES
(4, 'SF Bay Animal Care', 'sfbay-sanctuary', 2);
INSERT INTO WorkingGroups VALUES
(5, 'Press', 'press', 1);
