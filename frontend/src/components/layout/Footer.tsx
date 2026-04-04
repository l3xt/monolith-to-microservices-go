import { BookOpen } from "lucide-react";
import { ServiceStatus } from "./ServiceStatus";

export function Footer() {
  return (
    <footer className="border-t py-6 mt-auto">
      <div className="container flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2 text-muted-foreground">
          <BookOpen className="h-4 w-4" />
          <span className="text-sm">Microservices by L3XT</span>
        </div>

        <ServiceStatus />

        <a href="https://github.com/l3xt/monolith-to-microservices-go" className="text-sm text-muted-foreground">Ссылка на GitHub</a>
      </div>
    </footer>
  );
}
