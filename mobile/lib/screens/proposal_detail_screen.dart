import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/proposal.dart';
import '../models/vote.dart';
import '../providers/dao_provider.dart';

class ProposalDetailScreen extends StatefulWidget {
  final Proposal proposal;

  const ProposalDetailScreen({
    super.key,
    required this.proposal,
  });

  @override
  State<ProposalDetailScreen> createState() => _ProposalDetailScreenState();
}

class _ProposalDetailScreenState extends State<ProposalDetailScreen> {
  VoteChoice? _selectedChoice;
  final _reasonController = TextEditingController();
  final _weightController = TextEditingController();

  @override
  void dispose() {
    _reasonController.dispose();
    _weightController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Proposal Details'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildProposalHeader(),
            const SizedBox(height: 16),
            _buildProposalDetails(),
            const SizedBox(height: 16),
            _buildVotingResults(),
            const SizedBox(height: 16),
            _buildVotingSection(),
          ],
        ),
      ),
    );
  }

  Widget _buildProposalHeader() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.proposal.title,
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                _buildStatusChip(widget.proposal.status),
                const SizedBox(width: 8),
                _buildTypeChip(widget.proposal.type),
                const SizedBox(width: 8),
                _buildVotingTypeChip(widget.proposal.votingType),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              widget.proposal.description,
              style: Theme.of(context).textTheme.bodyLarge,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildProposalDetails() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Proposal Details',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            _buildDetailRow('ID', widget.proposal.id),
            _buildDetailRow('Creator', _truncateAddress(widget.proposal.creator)),
            _buildDetailRow('Start Time', _formatDateTime(widget.proposal.startTime)),
            _buildDetailRow('End Time', _formatDateTime(widget.proposal.endTime)),
            _buildDetailRow('Threshold', '${widget.proposal.threshold}'),
            if (widget.proposal.metadataHash != null)
              _buildDetailRow('Metadata Hash', _truncateAddress(widget.proposal.metadataHash!)),
          ],
        ),
      ),
    );
  }

  Widget _buildVotingResults() {
    final results = widget.proposal.results;
    if (results == null) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Voting Results',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              const Text('No votes cast yet'),
            ],
          ),
        ),
      );
    }

    final totalVotes = results.yesVotes + results.noVotes + results.abstainVotes;
    
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Voting Results',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                Icon(
                  results.passed ? Icons.check_circle : Icons.cancel,
                  color: results.passed ? Colors.green : Colors.red,
                ),
              ],
            ),
            const SizedBox(height: 16),
            
            // Yes Votes
            _buildVoteBar('Yes', results.yesVotes, totalVotes, Colors.green),
            const SizedBox(height: 8),
            
            // No Votes
            _buildVoteBar('No', results.noVotes, totalVotes, Colors.red),
            const SizedBox(height: 8),
            
            // Abstain Votes
            _buildVoteBar('Abstain', results.abstainVotes, totalVotes, Colors.grey),
            const SizedBox(height: 16),
            
            // Summary
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Total Voters: ${results.totalVoters}'),
                Text('Quorum: ${results.quorum}'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVoteBar(String label, int votes, int total, Color color) {
    final percentage = total > 0 ? votes / total : 0.0;
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label),
            Text('$votes (${(percentage * 100).toStringAsFixed(1)}%)'),
          ],
        ),
        const SizedBox(height: 4),
        LinearProgressIndicator(
          value: percentage,
          backgroundColor: Colors.grey[300],
          valueColor: AlwaysStoppedAnimation<Color>(color),
        ),
      ],
    );
  }

  Widget _buildVotingSection() {
    return Consumer<DAOProvider>(
      builder: (context, daoProvider, child) {
        if (!daoProvider.isWalletConnected) {
          return Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                children: [
                  Icon(Icons.wallet, size: 48, color: Colors.grey[400]),
                  const SizedBox(height: 16),
                  const Text('Connect your wallet to vote'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      // Navigate to wallet screen
                      Navigator.pushNamed(context, '/wallet');
                    },
                    child: const Text('Connect Wallet'),
                  ),
                ],
              ),
            ),
          );
        }

        if (widget.proposal.status != ProposalStatus.active) {
          return Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                children: [
                  Icon(Icons.how_to_vote_outlined, size: 48, color: Colors.grey[400]),
                  const SizedBox(height: 16),
                  Text('Voting is ${widget.proposal.status.name}'),
                ],
              ),
            ),
          );
        }

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Cast Your Vote',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 16),
                
                // Vote Choice Selection
                Text('Your Choice:', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: RadioListTile<VoteChoice>(
                        title: const Text('Yes'),
                        value: VoteChoice.yes,
                        groupValue: _selectedChoice,
                        onChanged: (value) {
                          setState(() {
                            _selectedChoice = value;
                          });
                        },
                      ),
                    ),
                    Expanded(
                      child: RadioListTile<VoteChoice>(
                        title: const Text('No'),
                        value: VoteChoice.no,
                        groupValue: _selectedChoice,
                        onChanged: (value) {
                          setState(() {
                            _selectedChoice = value;
                          });
                        },
                      ),
                    ),
                    Expanded(
                      child: RadioListTile<VoteChoice>(
                        title: const Text('Abstain'),
                        value: VoteChoice.abstain,
                        groupValue: _selectedChoice,
                        onChanged: (value) {
                          setState(() {
                            _selectedChoice = value;
                          });
                        },
                      ),
                    ),
                  ],
                ),
                
                const SizedBox(height: 16),
                
                // Vote Weight (for quadratic voting)
                if (widget.proposal.votingType == VotingType.quadratic) ...[
                  TextField(
                    controller: _weightController,
                    decoration: const InputDecoration(
                      labelText: 'Vote Weight',
                      border: OutlineInputBorder(),
                      helperText: 'Higher weight costs more tokens (quadratic)',
                    ),
                    keyboardType: TextInputType.number,
                  ),
                  const SizedBox(height: 16),
                ],
                
                // Vote Reason
                TextField(
                  controller: _reasonController,
                  decoration: const InputDecoration(
                    labelText: 'Reason (Optional)',
                    border: OutlineInputBorder(),
                    helperText: 'Explain your vote to the community',
                  ),
                  maxLines: 3,
                ),
                
                const SizedBox(height: 16),
                
                // Submit Vote Button
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _selectedChoice != null && !daoProvider.isLoading
                        ? _submitVote
                        : null,
                    child: daoProvider.isLoading
                        ? const Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              ),
                              SizedBox(width: 8),
                              Text('Submitting Vote...'),
                            ],
                          )
                        : const Text('Submit Vote'),
                  ),
                ),
              ],
            ),
          ),
        );
      },
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

  Widget _buildVotingTypeChip(VotingType votingType) {
    return Chip(
      label: Text(
        votingType.name.toUpperCase(),
        style: const TextStyle(fontSize: 10),
      ),
      backgroundColor: Colors.blue[100],
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(
            child: Text(value),
          ),
        ],
      ),
    );
  }

  String _truncateAddress(String address) {
    if (address.length <= 16) return address;
    return '${address.substring(0, 8)}...${address.substring(address.length - 8)}';
  }

  String _formatDateTime(DateTime dateTime) {
    return '${dateTime.day}/${dateTime.month}/${dateTime.year} ${dateTime.hour}:${dateTime.minute.toString().padLeft(2, '0')}';
  }

  Future<void> _submitVote() async {
    if (_selectedChoice == null) return;

    int weight = 1;
    if (widget.proposal.votingType == VotingType.quadratic) {
      weight = int.tryParse(_weightController.text) ?? 1;
    }

    try {
      await context.read<DAOProvider>().castVote(
        proposalId: widget.proposal.id,
        choice: _selectedChoice!,
        weight: weight,
        reason: _reasonController.text.isNotEmpty ? _reasonController.text : null,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Vote submitted successfully!')),
        );
        
        // Clear the form
        setState(() {
          _selectedChoice = null;
        });
        _reasonController.clear();
        _weightController.clear();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to submit vote: $e')),
        );
      }
    }
  }
}