import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WailsService } from '../../services/wails.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './header.component.html',
  styleUrl: './header.component.css'
})
export class HeaderComponent implements OnInit {
  version = 'Loading GNSSGO version...';
  logoPath = 'assets/images/logo-universal.png';

  constructor(private wailsService: WailsService) {}

  ngOnInit(): void {
    this.loadGNSSGOVersion();
  }

  async loadGNSSGOVersion(): Promise<void> {
    try {
      const version = await this.wailsService.getGNSSGOVersion();
      this.version = `GNSSGO Version: ${version}`;
    } catch (err) {
      console.error(err);
      this.version = 'Error loading GNSSGO version';
    }
  }
}
