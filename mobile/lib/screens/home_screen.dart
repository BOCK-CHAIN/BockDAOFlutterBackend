import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/dao_provider.dart';
import '../models/proposal.dart';
import 'proposal_detail_screen.dart';
import 'create_proposal_screen.dart';
import 'wallet_screen.dart';
import 'treasury_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedIndex = 0;
  ProposalStatus? _selectedStatus;
  ProposalType? _selectedType;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<DAOProvider>().initialize();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Bock DAO'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          Consumer<DAOProvider>(
            builder: (context, daoProvider, child) {
              return IconButton(
                icon: Icon(
                  daoProvider.isWalletConnected 
                    ? Icons.account_balance_wallet 
                    : Icons.wallet,
                  color: daoProvider.isWalletConnected 
                    ? Colors.green 
                    : Colors.grey,
                ),
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const WalletScreen()),
                  );
                },
              );
            },
          ),
        ],
      ),
      body: _buildBody(),
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        currentIndex: _selectedIndex,
        onTap: (index) {
          setState(() {
            _selectedIndex = index;
          });
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.home),
            label: 'Proposals',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.add_circle_outline),
            label: 'Create',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.account_balance),
            label: 'Treasury',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: 'Profile',
          ),
        ],
      ),
    );
  }

  Widget _buildBody() {
    switch (_selectedIndex) {
      case 0:
        return _buildProposalsTab();
      case 1:
        return const CreateProposalScreen();
      case 2:
        return const TreasuryScreen();
      case 3:
        return const WalletScreen();
      default:
        return _buildProposalsTab();
    }
  }

  Widget _buildProposalsTab() {
    return Column(
      children: [
        _buildFilters(),
        Expanded(child: _buildProposalsList()),
      ],
    );
  }

  Widget _buildFilters() {
    return Container(
      padding: const EdgeInsets.all(16.0),
      child: Row(
        children: [
          Expanded(
            child: DropdownButtonFormField<ProposalStatus>(
              decoration: const InputDecoration(
                labelText: 'Status',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              ),
              value: _selectedStatus,
              items: [
                const DropdownMenuItem<ProposalStatus>(
                  value: null,
                  child: Text('All Statuses'),
                ),
                ...ProposalStatus.values.map((status) => DropdownMenuItem(
                  value: status,
                  child: Text(status.name.toUpperCase()),
                )),
              ],
              onChanged: (value) {
                setState(() {
                  _selectedStatus = value;
                });
                _refreshProposals();
              },
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: DropdownButtonFormField<ProposalType>(
              decoration: const InputDecoration(
                labelText: 'Type',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              ),
              value: _selectedType,
              items: [
                const DropdownMenuItem<ProposalType>(
                  value: null,
                  child: Text('All Types'),
                ),
                ...ProposalType.values.map((type) => DropdownMenuItem(
                  value: type,
                  child: Text(type.name.toUpperCase()),
                )),
              ],
              onChanged: (value) {
                setState(() {
                  _selectedType = value;
                });
                _refreshProposals();
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProposalsList() {
    return Consumer<DAOProvider>(
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
                  onPressed: _refreshProposals,
                  child: const Text('Retry'),
                ),
              ],
            ),
          );
        }

        if (daoProvider.proposals.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.inbox, size: 64, color: Colors.grey[400]),
                const SizedBox(height: 16),
                Text(
                  'No proposals found',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  'Be the first to create a proposal!',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: _refreshProposals,
          child: ListView.builder(
            itemCount: daoProvider.proposals.length,
            itemBuilder: (context, index) {
              final proposal = daoProvider.proposals[index];
              return _buildProposalCard(proposal);
            },
          ),
        );
      },
    );
  }

  Widget _buildProposalCard(Proposal proposal) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: ListTile(
        title: Text(
          proposal.title,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text(
              proposal.description,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                _buildStatusChip(proposal.status),
                const SizedBox(width: 8),
                _buildTypeChip(proposal.type),
              ],
            ),
          ],
        ),
        trailing: const Icon(Icons.arrow_forward_ios),
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => ProposalDetailScreen(proposal: proposal),
            ),
          );
        },
      ),
    );
  }

  Widget _buildStatusChip(ProposalStatus status) {
    Color color;
    switch (status) {
      case ProposalStatus.active:
        color = Colors.green;
        break;
      case ProposalStatus.passed:
        color = Colors.blue;
        break;
      case ProposalStatus.rejected:
        color = Colors.red;
        break;
      case ProposalStatus.executed:
        color = Colors.purple;
        break;
      case ProposalStatus.cancelled:
        color = Colors.grey;
        break;
      default:
        color = Colors.orange;
    }

    return Chip(
      label: Text(
        status.name.toUpperCase(),
        style: const TextStyle(fontSize: 10, color: Colors.white),
      ),
      backgroundColor: color,
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }

  Widget _buildTypeChip(ProposalType type) {
    return Chip(
      label: Text(
        type.name.toUpperCase(),
        style: const TextStyle(fontSize: 10),
      ),
      backgroundColor: Colors.grey[200],
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }

  Future<void> _refreshProposals() async {
    await context.read<DAOProvider>().loadProposals(
      status: _selectedStatus,
      type: _selectedType,
    );
  }
}