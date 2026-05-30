import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'shared_card.dart';

class BrowserCard extends StatefulWidget {
  const BrowserCard({super.key});

  @override
  State<BrowserCard> createState() => _BrowserCardState();
}

class _BrowserCardState extends State<BrowserCard> {
  final TextEditingController _urlController = TextEditingController();
  bool _isLoading = false;

  Future<void> _openBrowser() async {
    String url = _urlController.text.trim();
    if (url.isEmpty || !url.contains('.')) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Masukkan domain yang valid, contoh: youtube.com'), backgroundColor: Colors.red),
      );
      return;
    }

    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      url = 'https://$url';
    }

    setState(() => _isLoading = true);
    final success = await ApiService.openBrowser(url);
    if (mounted) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? 'Dibuka di Edge' : 'Gagal membuka URL'),
          backgroundColor: success ? Colors.green : Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return SharedCard(
      child: Column(
        children: [
          const CardHeader(icon: Icons.open_in_browser, title: 'Browser (Edge)'),
          const SizedBox(height: 16),
          TextField(
            controller: _urlController,
            keyboardType: TextInputType.url,
            decoration: InputDecoration(
              hintText: 'contoh: youtube.com',
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: const Icon(Icons.clear),
                onPressed: () => _urlController.clear(),
              ),
            ),
            onChanged: (val) => setState(() {}),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              ActionChip(
                label: const Text('YouTube'),
                onPressed: () {
                  _urlController.text = 'youtube.com';
                  setState(() {});
                },
              ),
              const SizedBox(width: 8),
              ActionChip(
                label: const Text('Bersihkan'),
                onPressed: () {
                  _urlController.clear();
                  setState(() {});
                },
              ),
            ],
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            height: 40,
            child: ElevatedButton(
              onPressed: _urlController.text.isEmpty || _isLoading ? null : _openBrowser,
              style: ElevatedButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.primary,
                foregroundColor: Theme.of(context).colorScheme.onPrimary,
              ),
              child: _isLoading
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(color: Theme.of(context).colorScheme.onPrimary, strokeWidth: 2),
                    )
                  : const Text('Buka di Edge'),
            ),
          ),
        ],
      ),
    );
  }
}
