import { spawn } from 'child_process';
import { EventEmitter } from 'events';
import { resolveKodeBinary } from './gatekeeper';

export interface DaemonNotification {
  title: string;
  body: string;
  needsApproval: boolean;
  approvalPrompt?: string;
}

export class DaemonBridge extends EventEmitter {
  private child: ReturnType<typeof spawn> | null = null;
  private pendingApproval: boolean = false;

  constructor(private projectRoot: string) {
    super();
  }

  public start(pollInterval: number = 30) {
    if (this.child) {
      return;
    }

    const binary = resolveKodeBinary();
    
    // Run the daemon. The daemon prints specifically formatted boxes for notifications.
    this.child = spawn(binary, ['daemon', '--poll', pollInterval.toString()], {
      cwd: this.projectRoot,
      env: { ...process.env, KODE_DAEMON_IPC: '1' },
    });

    let buffer = '';

    this.child.stderr?.on('data', (data: Buffer) => {
      const text = data.toString();
      buffer += text;

      // Extract Kode Daemon notifications
      if (buffer.includes('┌─') && buffer.includes('└─')) {
        const matches = buffer.match(/┌─(.*?)─*┐\n([\s\S]*?)└─+─*┘/g);
        if (matches) {
          for (const match of matches) {
            const lines = match.split('\n');
            const title = lines[0].replace(/┌─|─*┐/g, '').trim();
            const body = lines.slice(1, -1)
              .map(l => l.replace(/│/g, '').trim())
              .filter(l => l.length > 0)
              .join('\n');
            
            this.emit('notification', { title, body, needsApproval: false });
          }
          // Clear buffer up to last match
          const lastMatchIndex = buffer.lastIndexOf('└─');
          buffer = buffer.substring(lastMatchIndex + 50); // Rough skip
        }
      }

      // Check for prompt
      if (text.includes('[Y/n]:')) {
        this.pendingApproval = true;
        const lines = text.split('\n');
        const promptLine = lines.find(l => l.includes('[Y/n]:')) || '';
        const promptClean = promptLine.replace('[Y/n]:', '').trim();
        
        this.emit('notification', { 
          title: 'Action Required', 
          body: promptClean, 
          needsApproval: true,
          approvalPrompt: promptClean
        });
      }
    });

    this.child.on('close', (code) => {
      this.child = null;
      this.emit('stopped', code);
    });
  }

  public respondToPrompt(approve: boolean) {
    if (!this.child || !this.pendingApproval) {
      return;
    }
    
    const response = approve ? "y\n" : "n\n";
    this.child.stdin?.write(response);
    this.pendingApproval = false;
  }

  public stop() {
    if (this.child) {
      this.child.kill('SIGTERM');
      this.child = null;
    }
  }
}
