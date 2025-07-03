select id, name, email, phone, interest_date, chapter_id from activists where email='activist@example.org';

select id, chapter_id, email, name, phone, processed from form_interest where email='activist@example.org';
