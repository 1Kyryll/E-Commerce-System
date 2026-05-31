import { Card } from "@/components/ui/card";

export function ProfileCard({ name, email }: { name?: string; email?: string }) {
  return (
    <Card>
      <Card.Header>
        <h2 className="font-secondary text-lg text-secondary">Profile</h2>
      </Card.Header>
      <Card.Body className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div>
          <p className="text-xs uppercase tracking-wide text-sub-accent-1">Name</p>
          <p className="text-secondary">{name ?? "—"}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-wide text-sub-accent-1">Email</p>
          <p className="text-secondary">{email ?? "—"}</p>
        </div>
      </Card.Body>
    </Card>
  );
}
