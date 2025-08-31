import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/dao_provider.dart';

class TreasuryScreen extends StatefulWidget {
  const TreasuryScreen({super.key});

  @override
  State<TreasuryScreen> createState() => _TreasuryScreenState();
}

class _TreasuryScreenState extends State<TreasuryScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<DAOProvider>().loadTreasuryStatus();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Treasury'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              context.read<DAOProvider>().loadTreasuryStatus();
            },
          ),
        ],
      ),
      body: Consumer<DAOProvider>(
        builder: (context, daoProvider, child) {
          if (daoProvider.isLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (daoProvider.error != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.error, size: 64, color: Colors.red[300]),
                  const SizedBox(height: 16),
                  Text(
                    'Error: ${daoProvider.error}',
                    style: Theme.of(context).textTheme.bodyLarge,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      daoProvider.loadTreasuryStatus();
                    },
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          final treasuryStatus = daoProvider.treasuryStatus;
          if (treasuryStatus == null) {
            return const Center(
              child: Text('No treasury data available'),
            );
          }

          return RefreshIndicator(
            onRefresh: () => daoProvider.loadTreasuryStatus(),
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildTreasuryOverview(treasuryStatus),
                  const SizedBox(height: 16),
                  _buildSignerInfo(treasuryStatus),
                  const SizedBox(height: 16),
                  _buildPendingTransactions(treasuryStatus),
                  const SizedBox(height: 16),
                  _buildRecentTransactions(treasuryStatus),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildTreasuryOverview(Map<String, dynamic> treasuryStatus) {
    final balance = treasuryStatus['balance'] ?? 0;
    final totalDisbursed = treasuryStatus['totalDisbursed'] ?? 0;
    final totalTransactions = treasuryStatus['totalTransactions'] ?? 0;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Treasury Overview',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 16),
            
            // Current Balance
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.green[50],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.green[200]!),
              ),
              child: Row(
                children: [
                  Icon(Icons.account_balance, color: Colors.green[700], size: 32),
                  const SizedBox(width: 16),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '$balance',
                        style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: Colors.green[700],
                        ),
                      ),
                      Text(
                        'Current Balance',
                        style: TextStyle(color: Colors.green[700]),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 16),
            
            // Statistics Row
            Row(
              children: [
                Expanded(
                  child: _buildStatCard(
                    'Total Disbursed',
                    '$totalDisbursed',
                    Icons.trending_down,
                    Colors.blue,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _buildStatCard(
                    'Transactions',
                    '$totalTransactions',
                    Icons.receipt,
                    Colors.orange,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatCard(String title, String value, IconData icon, Color color) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Column(
        children: [
          Icon(icon, color: color, size: 24),
          const SizedBox(height: 8),
          Text(
            value,
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
          Text(
            title,
            style: TextStyle(
              fontSize: 12,
              color: color,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _buildSignerInfo(Map<String, dynamic> treasuryStatus) {
    final signers = treasuryStatus['signers'] as List<dynamic>? ?? [];
    final requiredSigs = treasuryStatus['requiredSignatures'] ?? 0;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Multi-Signature Configuration',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            
            Row(
              children: [
                Icon(Icons.group, color: Colors.blue[700]),
                const SizedBox(width: 8),
                Text('${signers.length} Authorized Signers'),
                const Spacer(),
                Icon(Icons.security, color: Colors.green[700]),
                const SizedBox(width: 8),
                Text('$requiredSigs Required Signatures'),
              ],
            ),
            
            const SizedBox(height: 16),
            
            // Signers List
            if (signers.isNotEmpty) ...[
              Text(
                'Authorized Signers:',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              ...signers.map((signer) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Row(
                  children: [
                    Icon(Icons.person, size: 16, color: Colors.grey[600]),
                    const SizedBox(width: 8),
                    Text(_truncateAddress(signer.toString())),
                  ],
                ),
              )),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildPendingTransactions(Map<String, dynamic> treasuryStatus) {
    final pendingTxs = treasuryStatus['pendingTransactions'] as List<dynamic>? ?? [];

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(
                  'Pending Transactions',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const Spacer(),
                if (pendingTxs.isNotEmpty)
                  Chip(
                    label: Text('${pendingTxs.length}'),
                    backgroundColor: Colors.orange[100],
                  ),
              ],
            ),
            const SizedBox(height: 16),
            
            if (pendingTxs.isEmpty) ...[
              Center(
                child: Column(
                  children: [
                    Icon(Icons.check_circle, size: 48, color: Colors.green[300]),
                    const SizedBox(height: 8),
                    const Text('No pending transactions'),
                  ],
                ),
              ),
            ] else ...[
              ...pendingTxs.map((tx) => _buildTransactionCard(tx, true)),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildRecentTransactions(Map<String, dynamic> treasuryStatus) {
    final recentTxs = treasuryStatus['recentTransactions'] as List<dynamic>? ?? [];

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Recent Transactions',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            
            if (recentTxs.isEmpty) ...[
              const Center(
                child: Text('No recent transactions'),
              ),
            ] else ...[
              ...recentTxs.take(5).map((tx) => _buildTransactionCard(tx, false)),
              if (recentTxs.length > 5) ...[
                const SizedBox(height: 8),
                Center(
                  child: TextButton(
                    onPressed: () {
                      // TODO: Navigate to full transaction history
                    },
                    child: Text('View All ${recentTxs.length} Transactions'),
                  ),
                ),
              ],
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildTransactionCard(Map<String, dynamic> tx, bool isPending) {
    final id = tx['id']?.toString() ?? '';
    final recipient = tx['recipient']?.toString() ?? '';
    final amount = tx['amount'] ?? 0;
    final purpose = tx['purpose']?.toString() ?? '';
    final signatures = tx['signatures'] as List<dynamic>? ?? [];
    final requiredSigs = tx['requiredSignatures'] ?? 0;
    final executed = tx['executed'] ?? false;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(8),
        color: isPending ? Colors.orange[50] : Colors.grey[50],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  purpose.isNotEmpty ? purpose : 'Treasury Transaction',
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              if (isPending)
                Chip(
                  label: Text('${signatures.length}/$requiredSigs'),
                  backgroundColor: Colors.orange[200],
                  materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                )
              else if (executed)
                const Icon(Icons.check_circle, color: Colors.green, size: 20),
            ],
          ),
          const SizedBox(height: 8),
          
          Row(
            children: [
              Icon(Icons.account_balance_wallet, size: 16, color: Colors.grey[600]),
              const SizedBox(width: 4),
              Text('$amount tokens'),
              const SizedBox(width: 16),
              Icon(Icons.person, size: 16, color: Colors.grey[600]),
              const SizedBox(width: 4),
              Expanded(child: Text(_truncateAddress(recipient))),
            ],
          ),
          
          if (id.isNotEmpty) ...[
            const SizedBox(height: 4),
            Row(
              children: [
                Icon(Icons.fingerprint, size: 16, color: Colors.grey[600]),
                const SizedBox(width: 4),
                Text(
                  _truncateAddress(id),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey[600],
                    fontFamily: 'monospace',
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  String _truncateAddress(String address) {
    if (address.length <= 16) return address;
    return '${address.substring(0, 8)}...${address.substring(address.length - 8)}';
  }
}