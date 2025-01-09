import { VueComponent } from "@/components/VueComponent";

export default function EditEventPage() {
  return (
    <div>
      {/* TODO(jh): get the event id from url (or new event), get chapter name from session info? */}
      <VueComponent name="event-edit" data={{ id: "1", chapter: "SF Bay" }} />
    </div>
  );
}
