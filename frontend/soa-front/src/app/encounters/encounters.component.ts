import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { EncountersService, Encounter } from '../services/encounters.service';
import { PositionService } from '../services/position.service';
import { AuthService } from '../auth/services/auth.service';
import { ToastService } from '../shared/toast/toast.service';
import { ConfirmDialogService } from '../shared/confirm-dialog/confirm-dialog.service';

@Component({
  selector: 'app-encounters',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './encounters.component.html',
  styleUrl: './encounters.component.css'
})
export class EncountersComponent implements OnInit {
  encounters: Encounter[] = [];
  loading = false;
  error = '';
  isGuide = false;
  showCreateForm = false;
  createForm: FormGroup;
  claimingId: string | null = null;

  constructor(
    private encountersService: EncountersService,
    private positionService: PositionService,
    private authService: AuthService,
    private toastService: ToastService,
    private confirmDialogService: ConfirmDialogService,
    private fb: FormBuilder,
    private cdr: ChangeDetectorRef
  ) {
    this.createForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      latitude: [null, Validators.required],
      longitude: [null, Validators.required],
      radiusMeters: [50, [Validators.required, Validators.min(5)]],
      reward: [''],
    });
  }

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    this.isGuide = user?.role === 'guide' || user?.role === 'admin';
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.encountersService.list().subscribe({
      next: (res) => {
        this.encounters = res.encounters;
        this.loading = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.error = 'Greška pri učitavanju izazova.';
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  toggleCreateForm(): void {
    this.showCreateForm = !this.showCreateForm;
  }

  useSimulatedPosition(): void {
    const position = this.positionService.getPosition();
    if (!position) {
      this.toastService.error('Nema sačuvane pozicije. Prvo je postavi na stranici "Trenutna lokacija".');
      return;
    }
    this.createForm.patchValue({ latitude: position.latitude, longitude: position.longitude });
  }

  createEncounter(): void {
    if (this.createForm.invalid) {
      this.createForm.markAllAsTouched();
      return;
    }
    this.encountersService.create(this.createForm.value).subscribe({
      next: () => {
        this.toastService.success('Izazov je kreiran.');
        this.showCreateForm = false;
        this.createForm.reset({ radiusMeters: 50 });
        this.load();
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri kreiranju izazova.');
      }
    });
  }

  claim(encounter: Encounter): void {
    const position = this.positionService.getPosition();
    if (!position) {
      this.toastService.error('Nema sačuvane pozicije. Prvo je postavi na stranici "Trenutna lokacija".');
      return;
    }

    this.claimingId = encounter.id;
    this.encountersService.claim(encounter.id, position.latitude, position.longitude).subscribe({
      next: () => {
        this.claimingId = null;
        this.toastService.success(`Osvojen izazov: ${encounter.name}!`);
        this.load();
      },
      error: (err) => {
        this.claimingId = null;
        this.toastService.error(err.error?.error || 'Predaleko si da bi preuzeo/la ovaj izazov.');
        this.cdr.detectChanges();
      }
    });
  }

  async remove(encounter: Encounter): Promise<void> {
    const confirmed = await this.confirmDialogService.confirm(`Obrisati izazov "${encounter.name}"?`);
    if (!confirmed) return;

    this.encountersService.delete(encounter.id).subscribe({
      next: () => {
        this.toastService.success('Izazov obrisan.');
        this.load();
      },
      error: () => this.toastService.error('Greška pri brisanju izazova.')
    });
  }
}
