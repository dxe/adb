# Local development
### Testing
From project root:
- `go test ./form_processor -v`
- Specific test: `go test ./form_processor -v TestProcessFormApplicationForNoMatchingActivist`

# ADB Form Processor Overview
Various forms for getting data into the ADB.

### How the process currently works
- When the form at dxe.io/apply is submitted, it inserts a row into the form_application table.
- When the form at dxe.io/checkin (and some other variations of it) is submitted, it insert rows to the form_interest table.
- When people sign petitions on our public website, it also inserts a row into the form_interest table if they are in the Bay Area.
- Then the Go "form processor" runs every N min within the Go program to process the new rows of those two tables and update the activists table accordingly.

### Processor Overview
- We first try to match the new form submission to an existing activist using their exact name. (We don't check email first in case people share an email address – sometimes people do this, especially older couples.)
- If no name is matched, then we try to match on email address. (But first we check that there is only one non-hidden activist that exists for this email address to avoid the issue mentioned above. If there is more than 1, then we generate an error so that someone can manually intervene).
- If no name or email matches an existing activist, then we insert a new row into the activists table.
- The "processed" row in the respective form table changes from 0 to 1 after the row is successfully processed.

### Updating an existing activist who filled out the application form
- We add data for the following fields to the record in the activists table ONLY if the field in the activists table is currently blank:
 - email
 - phone
 - location
 - dob
- If new data is provided for these fields, it should be concated with the existing data in the activists table:
 - dev_interest
 - referral_friends
 - referral_apply
 - referral_outlet
- The various "prospective" fields are updated based on the application type & activist's current level:
 - If the application type is "organizer" and the activist level is not already an "organizer", then we set prospect_organizer to true. (Note that the existing logic also includes some Senior Organizer stuff, but we no longer have that activist level so it can be ignored.)
 - If the application type is "chapter-member" and the activist level is "supporter" or "non-local", then we set prospect_chapter_member to true. (Note that the existing logic also includes some Circle Member stuff, but we no longer have that activist level so it can be ignored.)
 - I believe the "circle-member" application type is no longer used, so that can be ignored.
- These fields are updated no matter what when something is successfully processed:
 - dev_application_date (current timestamp)
 - dev_application_type (the type of application submitted)

### Updating an existing activist who filled out the interest form
- We add data for the following fields to the record in the activists table ONLY if the field in the activists table is currently blank:
 - email
 - phone
 - location
 - source
- If the form type is "Circle Interest Form", then we set circle_interest to true.
- These fields are updated no matter what when something is successfully processed:
 - interest_date (current timestamp)
- If new data is provided for these fields, it should be concated with the existing data in the activists table:
 - dev_interest
 - referral_friends
 - referral_apply
 - referral_outlet

### Inserting a new activist from the application or interest form
- All data from the form submission should be added to the new activist record.
- " (inserted by application, check for duplicate)" should be appended to the activist's name so that someone can merge the record later if it is a duplicate of someone else whose name or email didn't match.
