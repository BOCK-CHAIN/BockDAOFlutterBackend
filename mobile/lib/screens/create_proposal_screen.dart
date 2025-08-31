import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/proposal.dart';
import '../providers/dao_provider.dart';

class CreateProposalScreen extends StatefulWidget {
  const CreateProposalScreen({super.key});

  @override
  State<CreateProposalScreen> createState() => _CreateProposalScreenState();
}

class _CreateProposalScreenState extends State<CreateProposalScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _thresholdController = TextEditingController();
  
  ProposalType _selectedType = ProposalType.general;
  VotingType _selectedVotingType = VotingType.simple;
  DateTime _startTime = DateTime.now().add(const Duration(hours: 1));
  DateTime _endTime = DateTime.now().add(const Duration(days: 7));

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _thresholdController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Create Proposal'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Consumer<DAOProvider>(
        builder: (context, daoProvider, child) {
          if (!daoProvider.isWalletConnected) {
            return _buildWalletRequired();
          }
          
          return _buildCreateProposalForm(daoProvider);
        },
      ),
    );
  }

  Widget _buildWalletRequired() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.wallet, size: 64, color: Colors.grey[400]),
            const SizedBox(height: 16),
            Text(
              'Wallet Required',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              'You need to connect your wallet to create proposals.',
              style: Theme.of(context).textTheme.bodyMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, '/wallet');
              },
              child: const Text('Connect Wallet'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCreateProposalForm(DAOProvider daoProvider) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Form(
        key: _formKey,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Title Field
            TextFormField(
              controller: _titleController,
              decoration: const InputDecoration(
                labelText: 'Proposal Title *',
                border: OutlineInputBorder(),
                helperText: 'A clear, concise title for your proposal',
              ),
              validator: (value) {
                if (value == null || value.isEmpty) {
                  return 'Please enter a title';
                }
                if (value.length < 10) {
                  return 'Title must be at least 10 characters';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Description Field
            TextFormField(
              controller: _descriptionController,
              decoration: const InputDecoration(
                labelText: 'Description *',
                border: OutlineInputBorder(),
                helperText: 'Detailed description of your proposal',
              ),
              maxLines: 5,
              validator: (value) {
                if (value == null || value.isEmpty) {
                  return 'Please enter a description';
                }
                if (value.length < 50) {
                  return 'Description must be at least 50 characters';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Proposal Type
            DropdownButtonFormField<ProposalType>(
              value: _selectedType,
              decoration: const InputDecoration(
                labelText: 'Proposal Type',
                border: OutlineInputBorder(),
              ),
              items: ProposalType.values.map((type) {
                return DropdownMenuItem(
                  value: type,
                  child: Text(_getProposalTypeLabel(type)),
                );
              }).toList(),
              onChanged: (value) {
                setState(() {
                  _selectedType = value!;
                });
              },
            ),
            const SizedBox(height: 16),

            // Voting Type
            DropdownButtonFormField<VotingType>(
              value: _selectedVotingType,
              decoration: const InputDecoration(
                labelText: 'Voting Type',
                border: OutlineInputBorder(),
              ),
              items: VotingType.values.map((type) {
                return DropdownMenuItem(
                  value: type,
                  child: Text(_getVotingTypeLabel(type)),
                );
              }).toList(),
              onChanged: (value) {
                setState(() {
                  _selectedVotingType = value!;
                });
              },
            ),
            const SizedBox(height: 16),

            // Threshold Field
            TextFormField(
              controller: _thresholdController,
              decoration: const InputDecoration(
                labelText: 'Voting Threshold',
                border: OutlineInputBorder(),
                helperText: 'Minimum votes required for proposal to pass',
              ),
              keyboardType: TextInputType.number,
              validator: (value) {
                if (value == null || value.isEmpty) {
                  return 'Please enter a threshold';
                }
                final threshold = int.tryParse(value);
                if (threshold == null || threshold <= 0) {
                  return 'Please enter a valid positive number';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Start Time
            Card(
              child: ListTile(
                title: const Text('Start Time'),
                subtitle: Text(_formatDateTime(_startTime)),
                trailing: const Icon(Icons.calendar_today),
                onTap: () => _selectDateTime(context, true),
              ),
            ),
            const SizedBox(height: 8),

            // End Time
            Card(
              child: ListTile(
                title: const Text('End Time'),
                subtitle: Text(_formatDateTime(_endTime)),
                trailing: const Icon(Icons.calendar_today),
                onTap: () => _selectDateTime(context, false),
              ),
            ),
            const SizedBox(height: 24),

            // Voting Type Information
            _buildVotingTypeInfo(),
            const SizedBox(height: 24),

            // Submit Button
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: daoProvider.isLoading ? null : _submitProposal,
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
                          Text('Creating Proposal...'),
                        ],
                      )
                    : const Text('Create Proposal'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVotingTypeInfo() {
    String info;
    switch (_selectedVotingType) {
      case VotingType.simple:
        info = 'Simple majority voting - each token holder gets one vote weighted by their token balance.';
        break;
      case VotingType.quadratic:
        info = 'Quadratic voting - voters can allocate multiple votes but the cost increases quadratically.';
        break;
      case VotingType.weighted:
        info = 'Token-weighted voting - voting power is directly proportional to token holdings.';
        break;
      case VotingType.reputation:
        info = 'Reputation-based voting - voting power is based on community reputation scores.';
        break;
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.blue[50],
        border: Border.all(color: Colors.blue[200]!),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Icon(Icons.info, color: Colors.blue[700]),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              info,
              style: TextStyle(color: Colors.blue[700]),
            ),
          ),
        ],
      ),
    );
  }

  String _getProposalTypeLabel(ProposalType type) {
    switch (type) {
      case ProposalType.general:
        return 'General Governance';
      case ProposalType.treasury:
        return 'Treasury Spending';
      case ProposalType.technical:
        return 'Technical Changes';
      case ProposalType.parameter:
        return 'Parameter Updates';
    }
  }

  String _getVotingTypeLabel(VotingType type) {
    switch (type) {
      case VotingType.simple:
        return 'Simple Majority';
      case VotingType.quadratic:
        return 'Quadratic Voting';
      case VotingType.weighted:
        return 'Token Weighted';
      case VotingType.reputation:
        return 'Reputation Based';
    }
  }

  String _formatDateTime(DateTime dateTime) {
    return '${dateTime.day}/${dateTime.month}/${dateTime.year} ${dateTime.hour}:${dateTime.minute.toString().padLeft(2, '0')}';
  }

  Future<void> _selectDateTime(BuildContext context, bool isStartTime) async {
    final scaffoldMessenger = ScaffoldMessenger.of(context);
    
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: isStartTime ? _startTime : _endTime,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (pickedDate != null && mounted) {
      final TimeOfDay? pickedTime = await showTimePicker(
        context: context,
        initialTime: TimeOfDay.fromDateTime(isStartTime ? _startTime : _endTime),
      );

      if (pickedTime != null && mounted) {
        final newDateTime = DateTime(
          pickedDate.year,
          pickedDate.month,
          pickedDate.day,
          pickedTime.hour,
          pickedTime.minute,
        );

        setState(() {
          if (isStartTime) {
            _startTime = newDateTime;
            // Ensure end time is after start time
            if (_endTime.isBefore(_startTime)) {
              _endTime = _startTime.add(const Duration(days: 7));
            }
          } else {
            if (newDateTime.isAfter(_startTime)) {
              _endTime = newDateTime;
            } else {
              // Show error message - end time must be after start time
              scaffoldMessenger.showSnackBar(
                const SnackBar(content: Text('End time must be after start time')),
              );
            }
          }
        });
      }
    }
  }

  Future<void> _submitProposal() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    // Additional validation
    if (_startTime.isBefore(DateTime.now())) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Start time must be in the future')),
      );
      return;
    }

    if (_endTime.isBefore(_startTime)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('End time must be after start time')),
      );
      return;
    }

    try {
      await context.read<DAOProvider>().createProposal(
        title: _titleController.text,
        description: _descriptionController.text,
        type: _selectedType,
        votingType: _selectedVotingType,
        startTime: _startTime,
        endTime: _endTime,
        threshold: int.parse(_thresholdController.text),
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Proposal created successfully!')),
        );

        // Clear the form
        _titleController.clear();
        _descriptionController.clear();
        _thresholdController.clear();
        setState(() {
          _selectedType = ProposalType.general;
          _selectedVotingType = VotingType.simple;
          _startTime = DateTime.now().add(const Duration(hours: 1));
          _endTime = DateTime.now().add(const Duration(days: 7));
        });

        // Navigate back to proposals list
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to create proposal: $e')),
        );
      }
    }
  }
}